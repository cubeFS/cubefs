// Copyright 2015 The etcd Authors
// Modified work copyright 2018 The tiglabs Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package raft

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/cubefs/cubefs/util/exporter"

	"github.com/tiglabs/raft/logger"
	"github.com/tiglabs/raft/proto"
	"github.com/tiglabs/raft/util"
)

type respondFunc func(interface{}, error)

type proposal struct {
	cmdType proto.EntryType
	respond respondFunc
	data    []byte
}

type askRollback struct {
	index   uint64 // entry index to be rollback
	data    []byte
	respond respondFunc
}

type apply struct {
	term        uint64
	index       uint64
	respond     respondFunc
	command     interface{}
	readIndexes []*Future
}

// handle user's get log entries request
type entryRequest struct {
	future     *Future
	index      uint64
	maxSize    uint64
	onlyCommit bool
}

type softState struct {
	leader uint64
	term   uint64
}

type peerState struct {
	peers map[uint64]proto.Peer
	mu    sync.RWMutex
}

type monitorStatus struct {
	conErrCount    uint8
	replicasErrCnt map[uint64]uint8
}

func (s *peerState) change(c *proto.ConfChange) {
	s.mu.Lock()
	switch c.Type {
	case proto.ConfAddNode, proto.ConfAddLearner, proto.ConfPromoteLearner:
		s.peers[c.Peer.ID] = c.Peer
	case proto.ConfRemoveNode:
		delete(s.peers, c.Peer.ID)
	case proto.ConfUpdateNode:
		s.peers[c.Peer.ID] = c.Peer
	}
	s.mu.Unlock()
}

func (s *peerState) replace(peers []proto.Peer) {
	s.mu.Lock()
	s.peers = nil
	s.peers = make(map[uint64]proto.Peer)
	for _, p := range peers {
		s.peers[p.ID] = p
	}
	s.mu.Unlock()
}

func (s *peerState) reset(peers []proto.Peer) {
	s.mu.Lock()
	//get old peer info
	newPeers := make([]proto.Peer, 0)
	for _, p := range peers {
		peer, ok := s.peers[p.ID]
		if ok {
			newPeers = append(newPeers, peer)
		}
	}

	//clear
	s.peers = nil

	//store exist old peer info
	s.peers = make(map[uint64]proto.Peer)
	for _, p := range newPeers {
		s.peers[p.ID] = p
	}
	s.mu.Unlock()
}

func (s *peerState) get() (nodes []uint64) {
	s.mu.RLock()
	for n := range s.peers {
		nodes = append(nodes, n)
	}
	s.mu.RUnlock()
	return
}

type pending struct {
	respond respondFunc
	typ     proto.EntryType
}

type raft struct {
	raftFsm           *raftFsm
	config            *Config
	raftConfig        *RaftConfig
	restoringSnapshot util.AtomicBool
	curApplied        util.AtomicUInt64
	curSoftSt         unsafe.Pointer
	prevSoftSt        softState
	prevHardSt        proto.HardState
	peerState         peerState
	pending           map[uint64]*pending
	snapping          map[uint64]*snapshotStatus
	mStatus           *monitorStatus
	propc             chan *proposal
	applyc            chan *apply
	askRollbackc      chan *askRollback
	recvc             chan *proto.Message
	snapRecvc         chan *snapshotRequest
	truncatec         chan uint64
	flushc            chan *Future
	readIndexC        chan *Future
	statusc           chan chan *Status
	entryRequestC     chan *entryRequest
	readyc            chan struct{}
	propReadyc        chan struct{}
	tickc             chan struct{}
	electc            chan struct{}
	promtec           chan struct{}
	stopc             chan struct{}
	done              chan struct{}
	mu                sync.Mutex
}

func newRaft(config *Config, raftConfig *RaftConfig) (*raft, error) {
	defer util.HandleCrash(fmt.Sprintf("newRaft[%v]", raftConfig.ID))

	if err := raftConfig.validate(); err != nil {
		return nil, err
	}

	r, err := newRaftFsm(config, raftConfig)
	if err != nil {
		return nil, err
	}

	mStatus := &monitorStatus{
		conErrCount:    0,
		replicasErrCnt: make(map[uint64]uint8),
	}
	raft := &raft{
		raftFsm:       r,
		config:        config,
		raftConfig:    raftConfig,
		mStatus:       mStatus,
		pending:       make(map[uint64]*pending),
		snapping:      make(map[uint64]*snapshotStatus),
		recvc:         make(chan *proto.Message, config.ReqBufferSize),
		applyc:        make(chan *apply, config.AppBufferSize),
		propc:         make(chan *proposal, 256),
		askRollbackc:  make(chan *askRollback, 256),
		snapRecvc:     make(chan *snapshotRequest, 1),
		truncatec:     make(chan uint64, 1),
		flushc:        make(chan *Future, 1),
		readIndexC:    make(chan *Future, 256),
		statusc:       make(chan chan *Status, 1),
		entryRequestC: make(chan *entryRequest, 16),
		tickc:         make(chan struct{}, 64),
		promtec:       make(chan struct{}, 64),
		readyc:        make(chan struct{}, 1),
		propReadyc:    make(chan struct{}, 1),
		electc:        make(chan struct{}, 1),
		stopc:         make(chan struct{}),
		done:          make(chan struct{}),
	}
	raft.raftFsm.registerRiskStateListener(raft.onFSMRiskStateChange)
	raft.raftFsm.registerAskRollbackListener(raft.onFSMAskRollback)
	raft.curApplied.Set(r.raftLog.applied)
	raft.peerState.replace(raftConfig.Peers)

	util.RunWorker(fmt.Sprintf("raft[%v]->runApply", r.id), raft.runApply, raft.handlePanic)
	util.RunWorker(fmt.Sprintf("raft[%v]->run", r.id), raft.run, raft.handlePanic)
	//util.RunWorker(raft.monitor, raft.handlePanic)
	return raft, nil
}

func (s *raft) stop() {
	select {
	case <-s.done:
		return
	default:
		s.doStop()
	}
	<-s.done

}

func (s *raft) doStop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.stopc:
		return
	default:
		s.raftFsm.StopFsm()
		close(s.stopc)
		s.restoringSnapshot.Set(false)
	}
}

func (s *raft) runApply() {
	defer func() {
		s.doStop()
		s.resetApply()
	}()

	var (
		loopCount = 0
		props     []*proposal
	)

	// This channel is used to receive signals indicating
	// that the raft.propc channel is currently available.
	var propReadyc <-chan struct{} = nil

	var propose = func(pr *proposal) {
		select {
		case s.propc <- pr:
		default:
			props = append(props, pr)
		}
	}

	for {
		loopCount = loopCount + 1
		if loopCount > 16 {
			loopCount = 0
			runtime.Gosched()
		}

		if len(props) > 0 {
			propReadyc = s.propReadyc
		}

		select {
		case <-s.stopc:
			return

		case askRollback := <-s.askRollbackc:

			if askRollback.index == 0 && len(askRollback.data) == 0 {
				// This is a committing resume notification.
				var proposal = pool.getProposal()
				proposal.cmdType = proto.EntryRollback
				proposal.data = nil
				proposal.respond = nil
				propose(proposal)
				continue
			}

			var (
				command []byte
				err     error
			)
			command, err = s.raftConfig.StateMachine.AskRollback(askRollback.data)

			if err != nil {
				logger.Warn("raft[%v] ask rollback for entry [index: %v], FSM returns error: %v", s.config.NodeID, askRollback.index, err)
				if askRollback.respond != nil {
					askRollback.respond(nil, err)
				}
				continue
			}
			if len(command) == 0 {
				if logger.IsEnableDebug() {
					logger.Debug("raft[%v] ask rollback for entry [index: %v], FSM returns empty command", s.config.NodeID, askRollback.index)
				}
				if askRollback.respond != nil {
					askRollback.respond(nil, nil)
				}
				continue
			}
			if logger.IsEnableDebug() {
				logger.Debug("raft[%v] ask rollback for entry [index: %v], FSM returns valid command", s.config.NodeID, askRollback.index)
			}
			var rollback = new(proto.Rollback)
			rollback.Index = askRollback.index
			rollback.Data = command

			var proposal = pool.getProposal()
			proposal.cmdType = proto.EntryRollback
			proposal.data = rollback.Encode()
			proposal.respond = askRollback.respond
			propose(proposal)

		case <-propReadyc:
			var (
				i         int
				breakLoop bool
			)
			for i < len(props) {
				select {
				case s.propc <- props[i]:
					i++
				default:
					breakLoop = true
				}
				if breakLoop {
					break
				}
			}
			switch {
			case i == 0:
			case i < len(props):
				props = append([]*proposal{}, props[i:]...)
			default:
				props = props[:0]
			}
			propReadyc = nil

		case apply := <-s.applyc:
			if apply.index <= s.curApplied.Get() {
				if len(apply.readIndexes) > 0 {
					respondReadIndex(apply.readIndexes, nil)
				}
				continue
			}

			var (
				err  error
				resp interface{}
			)
			switch cmd := apply.command.(type) {
			case *proto.ConfChange:
				resp, err = s.raftConfig.StateMachine.ApplyMemberChange(cmd, apply.index)
			case []byte:
				resp, err = s.raftConfig.StateMachine.Apply(cmd, apply.index)
			}

			if apply.respond != nil {
				apply.respond(resp, err)
			}
			if len(apply.readIndexes) > 0 {
				respondReadIndex(apply.readIndexes, nil)
			}
			s.curApplied.Set(apply.index)
			pool.returnApply(apply)

		}
	}
}

func (s *raft) run() {
	defer func() {
		s.doStop()
		s.resetPending(ErrStopped)
		s.raftFsm.readOnly.reset(ErrStopped)
		s.stopSnapping()
		s.raftConfig.Storage.Close()
		close(s.done)
	}()
	s.prevHardSt.Term = s.raftFsm.term
	s.prevHardSt.Vote = s.raftFsm.vote
	s.prevHardSt.Commit = s.raftFsm.raftLog.committed
	s.maybeChange(true)
	loopCount := 0
	var readyc chan struct{}
	for {
		if readyc == nil && s.containsUpdate() {
			readyc = s.readyc
			readyc <- struct{}{}
		}

		select {
		case <-s.stopc:
			return

		case <-s.tickc:
			s.raftFsm.tick()
			s.maybeChange(true)
		//case <-printTicker.C:
		//	var (
		//		getCnt,putCnt uint64
		//	)
		//	if s.isLeader(){
		//		getCnt=proto.LoadLeaderGetEntryCnt()
		//		putCnt=proto.LoadLeaderPutEntryCnt()
		//	}else {
		//		getCnt=proto.LoadFollowerGetEntryCnt()
		//		putCnt=proto.LoadFollowerPutEntryCnt()
		//	}
		//	fmt.Println(fmt.Sprintf("isLeaderRole(%v) getEntryCntFromPool(%v) putEntryCntFromPool(%v) ",s.isLeader(),getCnt,putCnt))
		case pr := <-s.propc:

			if s.raftFsm.leader != s.config.NodeID {
				if pr.respond != nil {
					pr.respond(nil, ErrNotLeader)
				}
				pool.returnProposal(pr)
				break
			}

			var msg = proto.GetMessage()
			msg.Type = proto.LocalMsgProp
			msg.From = s.config.NodeID

			var nextIndex = s.raftFsm.raftLog.lastIndex() + 1
			nextIndex, msg.Entries = s.handleProposal(pr, nextIndex, msg.Entries)

			const (
				maxBatchSize = 64
				maxLoopCount = 256
			)
			var (
				breakLoop    bool
				curLoopCount = 0
			)
			for len(msg.Entries) < maxBatchSize && curLoopCount < maxLoopCount {
				curLoopCount++
				select {
				case pr := <-s.propc:
					nextIndex, msg.Entries = s.handleProposal(pr, nextIndex, msg.Entries)
					continue
				default:
					breakLoop = true
				}
				if breakLoop {
					break
				}
			}

			s.markProposalChannelReady()

			if len(msg.Entries) == 0 {
				proto.ReturnMessage(msg)
				break
			}

			s.raftFsm.Step(msg)

		case m := <-s.recvc:
			// MsgFilter 仅用于单测中制造异常场景，正式代码中不要赋值！
			if s.raftConfig.MsgFilter(m) {
				proto.ReturnMessage(m)
				continue
			} else if _, ok := s.raftFsm.replicas[m.From]; ok || (!m.IsResponseMsg() && m.Type != proto.ReqMsgVote) ||
				(m.Type == proto.ReqMsgVote && s.raftFsm.raftLog.isUpToDate(m.Index, m.LogTerm, 0, 0)) {
				switch m.Type {
				case proto.ReqMsgHeartBeat:
					if s.raftFsm.leader == m.From && m.From != s.config.NodeID {
						s.raftFsm.Step(m)
					}
				case proto.RespMsgHeartBeat:
					if s.raftFsm.leader == s.config.NodeID && m.From != s.config.NodeID {
						s.raftFsm.Step(m)
					}
				default:
					s.raftFsm.Step(m)
				}
				var respErr = true
				if m.Type == proto.RespMsgAppend && m.Reject != true {
					respErr = false
				}
				s.maybeChange(respErr)
			} else if logger.IsEnableWarn() && m.Type != proto.RespMsgHeartBeat {
				logger.Warn("[raft][%v term: %d] ignored a %s message without the replica from [%v term: %d].", s.raftFsm.id, s.raftFsm.term, m.Type, m.From, m.Term)
			}

		case snapReq := <-s.snapRecvc:
			s.handleSnapshot(snapReq)

		case <-readyc:
			/*
				s.persist()

				s.apply()


				s.advance()

				// Send all messages.
				for _, msg := range s.raftFsm.msgs {
					if msg.Type == proto.ReqMsgSnapShot {
						s.sendSnapshot(msg)
						continue
					}
					s.sendMessage(msg)
				}
			*/
			if s.isLeader() {
				s.apply()
				for _, msg := range s.raftFsm.msgs {
					if msg.Type == proto.ReqMsgSnapShot {
						s.sendSnapshot(msg)
						continue
					}
					// MsgFilter 仅用于单测中制造异常场景，正式代码中不要赋值！
					if s.raftConfig.MsgFilter(msg) {
						proto.ReturnMessage(msg)
						continue
					}
					s.sendMessage(msg)
				}
				s.persist()
			} else {
				s.persist()
				for _, msg := range s.raftFsm.msgs {
					if msg.Type == proto.ReqMsgSnapShot {
						s.sendSnapshot(msg)
						continue
					}
					s.sendMessage(msg)
				}
				s.apply()
			}
			s.advance()

			s.raftFsm.msgs = nil
			readyc = nil
			loopCount = loopCount + 1
			if loopCount >= 2 {
				loopCount = 0
				runtime.Gosched()
			}

		case <-s.electc:
			msg := proto.GetMessage()
			msg.Type = proto.LocalMsgHup
			msg.From = s.config.NodeID
			msg.ForceVote = true
			s.raftFsm.Step(msg)
			s.maybeChange(true)

		case c := <-s.statusc:
			c <- s.getStatus()
		case f := <-s.flushc:
			err := s.raftConfig.Storage.Flush()
			if err != nil {
				logger.Error("raft[%v] flush storage failed: %v", err)
			}
			f.respond(nil, err)
		case truncIndex := <-s.truncatec:
			func(truncateTo uint64) {
				defer util.HandleCrash(fmt.Sprintf("raft[%v]->truncateTo", s.raftFsm.id))

				if lasti, err := s.raftConfig.Storage.LastIndex(); err != nil {
					logger.Error("raft[%v] truncate failed to get last index from storage: %v", s.raftFsm.id, err)
				} else if lasti > s.config.RetainLogs {
					truncateTo = util.Min(truncateTo, lasti-s.config.RetainLogs)
					var firsti uint64
					if firsti, err = s.raftConfig.Storage.FirstIndex(); err != nil {
						logger.Error("raft[%v] truncate failed to get first index from storage: %v", s.raftFsm.id, err)
						return
					}
					if truncateTo >= firsti {
						if err = s.raftConfig.Storage.Truncate(truncateTo); err != nil {
							logger.Error("raft[%v] truncate failed,error is: %v", s.raftFsm.id, err)
							return
						}
						logger.Debug("raft[%v] [firstindex: %v] truncate storage to %v", s.raftFsm.id, firsti, truncateTo)
					}

				}
			}(truncIndex)

		case <-s.promtec:
			s.promoteLearner()

		case future := <-s.readIndexC:
			futures := []*Future{future}
			// handle in batch
			var flag bool
			for i := 1; i < 64; i++ {
				select {
				case f := <-s.readIndexC:
					futures = append(futures, f)
				default:
					flag = true
				}
				if flag {
					break
				}
			}
			s.raftFsm.addReadIndex(futures)

		case req := <-s.entryRequestC:
			s.getEntriesInLoop(req)
		}
	}
}

func (s *raft) monitor() {
	statusTicker := time.NewTicker(5 * time.Second)
	leaderTicker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-s.stopc:
			statusTicker.Stop()
			leaderTicker.Stop()
			return

		case <-statusTicker.C:
			if s.raftFsm.leader == NoLeader || s.raftFsm.state == stateCandidate {
				s.mStatus.conErrCount++
			} else {
				s.mStatus.conErrCount = 0
			}
			if s.mStatus.conErrCount > 5 {
				errMsg := fmt.Sprintf("raft status not health partitionID[%d]_nodeID[%d]_leader[%v]_state[%v]_replicas[%v]",
					s.raftFsm.id, s.raftFsm.config.NodeID, s.raftFsm.leader, s.raftFsm.state, s.raftFsm.peers())
				exporter.Warning(errMsg)
				logger.Error(errMsg)

				s.mStatus.conErrCount = 0
			}
		case <-leaderTicker.C:
			if s.raftFsm.state == stateLeader {
				for id, p := range s.raftFsm.replicas {
					if id == s.raftFsm.config.NodeID {
						continue
					}
					if p.active == false {
						s.mStatus.replicasErrCnt[id]++
					} else {
						s.mStatus.replicasErrCnt[id] = 0
					}
					if s.mStatus.replicasErrCnt[id] > 5 {
						errMsg := fmt.Sprintf("raft partitionID[%d] replicaID[%v] not active peer[%v]",
							s.raftFsm.id, id, p.peer)
						exporter.Warning(errMsg)
						logger.Error(errMsg)
						s.mStatus.replicasErrCnt[id] = 0
					}
				}
			}
		}
	}
}

func (s *raft) tick() {
	if s.restoringSnapshot.Get() {
		return
	}

	select {
	case <-s.stopc:
	case s.tickc <- struct{}{}:
	default:
		return
	}
}

func (s *raft) promote() {
	if !s.isLeader() {
		return
	}
	select {
	case <-s.stopc:
	case s.promtec <- struct{}{}:
	default:
		return
	}
}

func (s *raft) propose(cmd []byte, future *Future) {

	if !s.isLeader() {
		future.respond(nil, ErrNotLeader)
		return
	}

	if !s.isAllowPropose() {
		future.respond(nil, ErrProposalDenied)
		return
	}

	pr := pool.getProposal()
	pr.cmdType = proto.EntryNormal
	pr.data = cmd
	pr.respond = future.respond

	select {
	case <-s.stopc:
		future.respond(nil, ErrStopped)
	case s.propc <- pr:
	}
}

func (s *raft) proposeMemberChange(cc *proto.ConfChange, future *Future) {
	if !s.isLeader() {
		future.respond(nil, ErrNotLeader)
		return
	}

	pr := pool.getProposal()
	pr.cmdType = proto.EntryConfChange
	pr.respond = future.respond
	pr.data = cc.Encode()

	select {
	case <-s.stopc:
		future.respond(nil, ErrStopped)
	case s.propc <- pr:
	}
}

// deal with 'ConfPromoteLearner' ConfChange
func (s *raft) proposePromoteLearnerMemberChange(cc *proto.ConfChange, future *Future, autoPromote bool) {
	if !s.isLeader() {
		future.respond(nil, ErrNotLeader)
		return
	}

	req := &proto.ConfChangeLearnerReq{}
	if err := json.Unmarshal(cc.Context, req); err != nil {
		logger.Error("raft[%v] json unmarshal ConfChangeLearnerReq Context[%s] err[%v]", s.raftFsm.id, string(cc.Context), err)
		future.respond(nil, ErrUnmarshal)
		return
	}
	replica, ok := s.raftFsm.replicas[cc.Peer.ID]
	if !ok || !s.isLearnerReady(replica, req.ChangeLearner.PromConfig.PromThreshold) {
		logger.Warn("raft[%v] promote learner[%s] err[%v]", s.raftFsm.id, string(cc.Context), ErrLearnerNotReady)
		future.respond(nil, ErrLearnerNotReady)
		return
	}

	pr := pool.getProposal()
	pr.cmdType = proto.EntryConfChange
	pr.respond = future.respond
	pr.data = cc.Encode()

	if autoPromote {
		select {
		case <-s.stopc:
			future.respond(nil, ErrStopped)
		case s.propc <- pr:
		default:
			logger.Warn("raft[%v] promote learner[%s] err[%v]", s.raftFsm.id, string(cc.Context), ErrFullChannel)
			future.respond(nil, ErrFullChannel)
			pool.returnProposal(pr)
		}
	} else {
		select {
		case <-s.stopc:
			future.respond(nil, ErrStopped)
		case s.propc <- pr:
		}
	}
	return
}

func (s *raft) setConsistencyMode(mode ConsistencyMode) {
	s.raftFsm.setConsistencyMode(mode)
}

func (s *raft) getConsistencyMode() ConsistencyMode {
	return s.raftFsm.getConsistencyMode()
}

func (s *raft) checkProposalACL(cmdType proto.EntryType) (allowed bool) {
	allowed = !(((s.raftFsm.getRiskStatus().Equals(UnstableState) && s.raftFsm.getConsistencyMode().Equals(StrictMode)) || s.raftFsm.isCommittingPaused()) && cmdType == proto.EntryNormal)
	return
}

func (s *raft) reciveMessage(m *proto.Message) {
	if s.restoringSnapshot.Get() {
		return
	}

	select {
	case <-s.stopc:
	case s.recvc <- m:
	default:
		logger.Warn(fmt.Sprintf("raft[%v] discard message(%v)", s.raftConfig.ID, m.ToString()))
		return
	}
}

func (s *raft) reciveSnapshot(m *snapshotRequest) {
	if s.restoringSnapshot.Get() {
		m.respond(ErrSnapping)
		return
	}

	select {
	case <-s.stopc:
		m.respond(ErrStopped)
		return
	case s.snapRecvc <- m:
	}
}

func (s *raft) status() *Status {
	if s.restoringSnapshot.Get() {
		return &Status{
			ID:                s.raftFsm.id,
			NodeID:            s.config.NodeID,
			RestoringSnapshot: true,
			State:             stateFollower.String(),
		}
	}

	c := make(chan *Status, 1)
	select {
	case <-s.stopc:
		return nil
	case s.statusc <- c:
		return <-c
	}
}

func (s *raft) truncate(index uint64) {
	if s.restoringSnapshot.Get() {
		return
	}

	select {
	case <-s.stopc:
	case s.truncatec <- index:
	default:
		return
	}
}

func (s *raft) flush(wait bool) (err error) {
	if s.restoringSnapshot.Get() {
		return
	}

	future := newFuture()

	select {
	case <-s.stopc:
	case s.flushc <- future:
		if wait {
			_, err = future.Response()
		}
	default:
	}
	return
}

func (s *raft) tryToLeader(future *Future) {
	if s.restoringSnapshot.Get() {
		future.respond(nil, nil)
		return
	}

	select {
	case <-s.stopc:
		future.respond(nil, ErrStopped)
	case s.electc <- struct{}{}:
		future.respond(nil, nil)
	}
}

func (s *raft) leaderTerm() (leader, term uint64) {
	st := (*softState)(atomic.LoadPointer(&s.curSoftSt))
	if st == nil {
		return NoLeader, 0
	}
	return st.leader, st.term
}

func (s *raft) isLeader() bool {
	leader, _ := s.leaderTerm()
	return leader == s.config.NodeID
}

func (s *raft) isAllowPropose() bool {
	return s.checkProposalACL(proto.EntryNormal)
}

func (s *raft) getRiskState() RiskState {
	return s.raftFsm.getRiskStatus()
}

func (s *raft) applied() uint64 {
	return s.curApplied.Get()
}

func (s *raft) committed() uint64 {
	return s.raftFsm.raftLog.committed
}

func (s *raft) sendMessage(m *proto.Message) {
	s.config.transport.Send(m)
}

func (s *raft) maybeChange(respErr bool) {
	updated := false
	if s.prevSoftSt.term != s.raftFsm.term {
		updated = true
		s.prevSoftSt.term = s.raftFsm.term
		s.resetTick()
	}
	preLeader := s.prevSoftSt.leader
	if preLeader != s.raftFsm.leader {
		updated = true
		s.prevSoftSt.leader = s.raftFsm.leader
		if s.raftFsm.leader != s.config.NodeID {
			if respErr == true || preLeader != s.config.NodeID {
				s.resetPending(ErrNotLeader)
			}
			s.stopSnapping()
		}
		if logger.IsEnableWarn() {
			if s.raftFsm.leader != NoLeader {
				if preLeader == NoLeader {
					logger.Warn("raft:[%v] elected leader %v at term %d.", s.raftFsm.id, s.raftFsm.leader, s.raftFsm.term)
				} else {
					logger.Warn("raft:[%v] changed leader from %v to %v at term %d.", s.raftFsm.id, preLeader, s.raftFsm.leader, s.raftFsm.term)
				}
			} else {
				logger.Warn("raft:[%v] lost leader %v at term %d.", s.raftFsm.id, preLeader, s.raftFsm.term)
			}
		}

		s.raftConfig.StateMachine.HandleLeaderChange(s.raftFsm.leader)
	}
	if updated {
		atomic.StorePointer(&s.curSoftSt, unsafe.Pointer(&softState{leader: s.raftFsm.leader, term: s.raftFsm.term}))
	}
}

func (s *raft) persist() {

	if err := s.raftFsm.raftLog.persist(); err != nil {
		panic(AppPanicError(fmt.Sprintf("[raft->persist][%v] storage storeEntries err: [%v].", s.raftFsm.id, err)))
	}
	if s.raftFsm.raftLog.committed != s.prevHardSt.Commit || s.raftFsm.term != s.prevHardSt.Term || s.raftFsm.vote != s.prevHardSt.Vote {
		hs := proto.HardState{Term: s.raftFsm.term, Vote: s.raftFsm.vote, Commit: s.raftFsm.raftLog.committed}
		if err := s.raftConfig.Storage.StoreHardState(hs); err != nil {
			panic(AppPanicError(fmt.Sprintf("[raft->persist][%v] storage storeHardState err: [%v].", s.raftFsm.id, err)))
		}
		s.prevHardSt = hs
	}

	if s.raftFsm.getRiskStatus().Equals(UnstableState) && s.config.SyncWALOnUnstable {
		if err := s.raftConfig.Storage.Flush(); err != nil {
			panic(AppPanicError(fmt.Sprintf("[raft->persist][%v] flush storage err: [%v].", s.raftFsm.id, err)))
		}
	}
}

func (s *raft) apply() {
	committedEntries := s.raftFsm.raftLog.nextEnts(noLimit)
	// check ready read index
	if len(committedEntries) == 0 {
		readIndexes := s.raftFsm.readOnly.getReady(s.curApplied.Get())
		if len(readIndexes) == 0 {
			return
		}
		apply := pool.getApply()
		apply.readIndexes = readIndexes
		select {
		case <-s.stopc:
			respondReadIndex(readIndexes, ErrStopped)
		case s.applyc <- apply:
		}
		return
	}

	for _, entry := range committedEntries {
		apply := pool.getApply()
		apply.term = entry.Term
		apply.index = entry.Index
		if pending, ok := s.pending[entry.Index]; ok {
			if pending.respond != nil {
				apply.respond = pending.respond
			}
			delete(s.pending, entry.Index)
			pool.returnPending(pending)
		}

		apply.readIndexes = s.raftFsm.readOnly.getReady(entry.Index)

		switch entry.Type {
		case proto.EntryNormal:
			if len(entry.Data) > 0 {
				apply.command = entry.Data
			}

		case proto.EntryConfChange:
			cc := new(proto.ConfChange)
			cc.Decode(entry.Data)
			apply.command = cc
			// repl apply
			s.raftFsm.applyConfChange(cc)
			s.peerState.change(cc)
			if logger.IsEnableWarn() {
				logger.Warn("raft[%v] applying configuration change [index: %v], detail: %v.", s.raftFsm.id, entry.Index, cc)
			}

		case proto.EntryRollback:
			rollback := new(proto.Rollback)
			rollback.Decode(entry.Data)
			if len(rollback.Data) > 0 {
				apply.command = rollback.Data
				if logger.IsEnableWarn() {
					logger.Warn("raft[%v] applying rollback entry [index: %v], rollback target [index: %v]", s.raftFsm.id, entry.Index, rollback.Index)
				}
			}
		}

		select {
		case <-s.stopc:
			if apply.respond != nil {
				apply.respond(nil, ErrStopped)
			}
			if len(apply.readIndexes) > 0 {
				respondReadIndex(apply.readIndexes, ErrStopped)
			}
		case s.applyc <- apply:
		}

	}
}

func (s *raft) advance() {
	s.raftFsm.raftLog.appliedTo(s.raftFsm.raftLog.committed)
}

func (s *raft) containsUpdate() bool {
	return len(s.raftFsm.raftLog.unstableEntries()) > 0 || s.raftFsm.raftLog.committed > s.raftFsm.raftLog.applied || len(s.raftFsm.msgs) > 0 ||
		s.raftFsm.raftLog.committed != s.prevHardSt.Commit || s.raftFsm.term != s.prevHardSt.Term || s.raftFsm.vote != s.prevHardSt.Vote ||
		s.raftFsm.readOnly.containsUpdate(s.curApplied.Get())
}

func (s *raft) resetPending(err error) {
	if len(s.pending) > 0 {
		for index, pending := range s.pending {
			if pending.respond != nil {
				pending.respond(nil, err)
			}
			delete(s.pending, index)
		}
	}
}

func (s *raft) onFSMAskRollback(startIndex uint64) (n int, err error) {
	var entries []*proto.Entry
	if entries, err = s.raftFsm.raftLog.entries(startIndex, noLimit); err != nil {
		return
	}
	var wrapRespond = func(respond respondFunc) respondFunc {
		if respond != nil {
			return func(i interface{}, err error) {
				if err != nil {
					respond(nil, ErrRollbackFailed)
					return
				}
				respond(nil, ErrProposalAbort)
				return
			}
		}
		return nil
	}

	var needRollbackEntries = filterNeedRollbackEntries(entries)
	if n = len(needRollbackEntries); n > 0 {
		var (
			from = needRollbackEntries[0].Index
			to   = needRollbackEntries[n-1].Index
		)
		for i := 0; i < n; i++ {
			var ent = needRollbackEntries[i]
			var respond respondFunc = nil
			if pending, exists := s.pending[ent.Index]; exists {
				delete(s.pending, ent.Index)
				respond = wrapRespond(pending.respond)
				pool.returnPending(pending)
			}
			s.askRollbackc <- &askRollback{
				index:   ent.Index,
				data:    ent.Data,
				respond: respond,
			}
		}
		// Signal for committing resume
		s.askRollbackc <- &askRollback{
			index:   0,
			data:    nil,
			respond: nil,
		}
		if logger.IsEnableWarn() {
			logger.Warn("raft[%v] prepare ask rollback for normal entries [from: %v, to: %v, num: %v]", s.raftConfig.ID, from, to, n)
		}
	}
	return
}

func (s *raft) resetTick() {
	for {
		select {
		case <-s.tickc:
		default:
			return
		}
	}
}

func (s *raft) resetApply() {
	for {
		select {
		case apply := <-s.applyc:
			if apply.respond != nil {
				apply.respond(nil, ErrStopped)
			}
			if len(apply.readIndexes) > 0 {
				respondReadIndex(apply.readIndexes, ErrStopped)
			}
			pool.returnApply(apply)
		default:
			return
		}
	}
}

func (s *raft) getStatus() *Status {
	stopped := false
	select {
	case <-s.stopc:
		stopped = true
	default:
	}

	st := &Status{
		ID:                s.raftFsm.id,
		NodeID:            s.config.NodeID,
		Leader:            s.raftFsm.leader,
		Term:              s.raftFsm.term,
		Index:             s.raftFsm.raftLog.lastIndex(),
		Commit:            s.raftFsm.raftLog.committed,
		Applied:           s.curApplied.Get(),
		Vote:              s.raftFsm.vote,
		State:             s.raftFsm.state.String(),
		RestoringSnapshot: s.restoringSnapshot.Get(),
		PendQueue:         len(s.pending),
		Pending: func() []PendingInfo {
			ret := make([]PendingInfo, 0, len(s.pending))
			for index, pending := range s.pending {
				ret = append(ret, PendingInfo{
					Index: index,
					Type:  pending.typ.String(),
				})
			}
			return ret
		}(),
		RecvQueue: len(s.recvc),
		AppQueue:  len(s.applyc),
		Stopped:   stopped,
		Log: LogStatus{
			FirstIndex: s.raftFsm.raftLog.firstIndex(),
			LastIndex:  s.raftFsm.raftLog.lastIndex(),
		},
		RistState: s.raftFsm.riskState.String(),
		Mode:      s.raftFsm.consistencyMode.String(),
	}
	if s.raftFsm.state == stateLeader {
		st.Replicas = make(map[uint64]*ReplicaStatus)
		for id, p := range s.raftFsm.replicas {
			st.Replicas[id] = &ReplicaStatus{
				Match:       p.match,
				Commit:      p.committed,
				Next:        p.next,
				State:       p.state.String(),
				Snapshoting: p.state == replicaStateSnapshot,
				Paused:      p.paused,
				Active:      p.active,
				LastActive:  p.lastActive,
				Inflight:    p.count,
				IsLearner:   p.isLearner,
				PromConfig:  p.promConfig,
			}
		}
	}
	return st
}

func (s *raft) handlePanic(err interface{}) {
	fatalStopc <- s.raftFsm.id

	fatal := &FatalError{
		ID:  s.raftFsm.id,
		Err: fmt.Errorf("raftID[%v] raft[%v] occur panic error: [%v]", s.config.NodeID, s.raftFsm.id, err),
	}
	s.raftConfig.StateMachine.HandleFatalEvent(fatal)
}

func (s *raft) getPeers() (peers []uint64) {
	return s.peerState.get()
}

func (s *raft) readIndex(future *Future) {
	if !s.isLeader() {
		future.respond(nil, ErrNotLeader)
		return
	}

	select {
	case <-s.stopc:
		future.respond(nil, ErrStopped)
	case s.readIndexC <- future:
	}
}

func (s *raft) getEntries(future *Future, startIndex uint64, maxSize uint64) {
	req := &entryRequest{
		future:  future,
		index:   startIndex,
		maxSize: maxSize,
	}
	select {
	case <-s.stopc:
		future.respond(nil, ErrStopped)
	case s.entryRequestC <- req:
	}
}

func (s *raft) getEntriesInLoop(req *entryRequest) {
	select {
	case <-s.stopc:
		req.future.respond(nil, ErrStopped)
		return
	default:
	}

	if !s.isLeader() {
		req.future.respond(nil, ErrNotLeader)
		return
	}
	if req.index > s.raftFsm.raftLog.lastIndex() {
		req.future.respond(nil, nil)
		return
	}
	if req.index < s.raftFsm.raftLog.firstIndex() {
		req.future.respond(nil, ErrCompacted)
		return
	}
	entries, err := s.raftFsm.raftLog.entries(req.index, req.maxSize)
	req.future.respond(entries, err)
}

func (s *raft) isLearnerReady(pr *replica, threshold uint8) bool {
	if !pr.isLearner || !pr.active {
		return false
	}
	leaderPr, ok := s.raftFsm.replicas[s.config.NodeID]
	if !ok || float64(pr.match) < float64(leaderPr.match)*float64(threshold)*0.01 {
		return false
	}
	// todo learner as quorum?
	if !s.raftFsm.checkLeaderLease(true) {
		return false
	}
	if logger.IsEnableDebug() {
		logger.Debug("raft[%v] leader[%v] promote learner[%v], leader match[%v], threshold[%v]",
			s.raftConfig.ID, s.config.NodeID, pr, leaderPr.match, threshold)
	}
	return true
}

func (s *raft) promoteLearner() {
	for id, pr := range s.raftFsm.replicas {
		if pr.isLearner && pr.promConfig.AutoPromote {
			future := newFuture()
			lear := proto.Learner{ID: id, PromConfig: &proto.PromoteConfig{AutoPromote: true, PromThreshold: pr.promConfig.PromThreshold}}
			req := &proto.ConfChangeLearnerReq{Id: s.raftConfig.ID, ChangeLearner: lear}
			bytes, err := json.Marshal(req)
			if err != nil {
				logger.Error("raft[%v] json marshal ConfChangeLearnerReq[%v] err[%v]", s.raftConfig.ID, req, err)
				continue
			}
			p := proto.Peer{ID: id}
			s.proposePromoteLearnerMemberChange(&proto.ConfChange{Type: proto.ConfPromoteLearner, Peer: p, Context: bytes}, future, true)
			//resp, err := future.Response()
			logger.Warn("raft[%v] leader[%v] auto promote learner[%v]", s.raftConfig.ID, s.config.NodeID, id)
		}
	}
}

func (s *raft) onFSMRiskStateChange(state RiskState) {
	if s.raftFsm.getRiskStatus().Equals(UnstableState) {
		if s.config.SyncWALOnUnstable {
			_ = s.flush(false)
			if logger.IsEnableDebug() {
				logger.Debug("raft[%v] proposed storage force flush cause risk state change to [%v].", s.raftConfig.ID, state)
			}
		}
	}
}

func (s *raft) handleProposal(pr *proposal, nextIndex uint64, entries []*proto.Entry) (uint64, []*proto.Entry) {
	switch {
	case s.maybeResumeCommitting(pr, nextIndex-1):
	case s.checkProposalACL(pr.cmdType):
		pending := pool.getPending()
		pending.typ = pr.cmdType
		pending.respond = pr.respond
		s.pending[nextIndex] = pending
		var e = &proto.Entry{Term: s.raftFsm.term, Index: nextIndex, Type: pr.cmdType, Data: pr.data}
		entries = append(entries, e)
		nextIndex++
	default:
		if pr.respond != nil {
			pr.respond(nil, ErrProposalDenied)
		}
	}
	pool.returnProposal(pr)
	return nextIndex, entries
}

func (s *raft) maybeResumeCommitting(pr *proposal, curMaxIndex uint64) bool {
	if pr.cmdType == proto.EntryRollback && len(pr.data) == 0 {
		s.raftFsm.setMinimumCommitIndex(curMaxIndex)
		if logger.IsEnableDebug() {
			logger.Debug("raft[%v] set minimum commit index to %v (auto resume committing), current committed %v", s.raftConfig.ID, curMaxIndex, s.raftFsm.raftLog.committed)
		}
		return true
	}
	return false
}

func (s *raft) markProposalChannelReady() {
	select {
	case s.propReadyc <- struct{}{}:
	default:
	}
}

// filterNeedRollbackEntries returns entries can be makes rollback from specified entries collection.
func filterNeedRollbackEntries(ents []*proto.Entry) []*proto.Entry {
	if len(ents) == 0 {
		return nil
	}
	var rollbacks = make(map[uint64]struct{})
	for i := 0; i < len(ents); i++ {
		if ents[i].Type == proto.EntryRollback {
			rollback := new(proto.Rollback)
			rollback.Decode(ents[i].Data)
			rollbacks[rollback.Index] = struct{}{}
		}
	}
	var result = make([]*proto.Entry, 0, len(ents))
	for i := 0; i < len(ents); i++ {
		if ent := ents[i]; ent.Type == proto.EntryNormal && len(ent.Data) > 0 {
			if _, exists := rollbacks[ent.Index]; exists {
				continue
			}
			result = append(result, ent)
		}
	}
	return result
}
