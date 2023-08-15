// Copyright 2018 The CubeFS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package datanode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/cubefs/cubefs/util/exporter"

	"github.com/cubefs/cubefs/proto"
	"github.com/cubefs/cubefs/util/log"
	"github.com/tiglabs/raft"
	raftproto "github.com/tiglabs/raft/proto"
)

/* The functions below implement the interfaces defined in the raft library. */

// Apply puts the data onto the disk.
func (dp *DataPartition) handleRaftApply(command []byte, index uint64) (resp interface{}, err error) {
	defer func() {
		if err != nil {
			msg := fmt.Sprintf("partition [id: %v, disk: %v] apply command [index: %v] occurred error and will be stop: %v",
				dp.partitionID, dp.Disk().Path, index, err)
			log.LogErrorf(msg)
			exporter.WarningCritical(msg)
			dp.Disk().space.DetachDataPartition(dp.partitionID)
			dp.Disk().DetachDataPartition(dp)
			dp.Stop()
			log.LogCriticalf("partition [id: %v, disk: %v] apply command [index: %v] failed and stopped: %v",
				dp.partitionID, dp.Disk().Path, index, err)
			return
		}
		dp.advanceApplyID(index)
		dp.actionHolder.Unregister(index)
	}()
	var opItem *rndWrtOpItem
	if opItem, err = UnmarshalRandWriteRaftLog(command); err != nil {
		resp = proto.OpErr
		return
	}
	resp, err = dp.ApplyRandomWrite(opItem, index)
	PutRandomWriteOpItem(opItem)
	return
}

// ApplyMemberChange supports adding new raft member or deleting an existing raft member.
// It does not support updating an existing member at this point.
func (dp *DataPartition) handleRaftApplyMemberChange(confChange *raftproto.ConfChange, index uint64) (resp interface{}, err error) {
	defer func(index uint64) {
		if err != nil {
			msg := fmt.Sprintf("partition [id: %v, disk: %v] apply member change [index: %v] occurred error and will be stop: %v",
				dp.partitionID, dp.Disk().Path, index, err)
			log.LogErrorf(msg)
			exporter.WarningCritical(msg)
			dp.Disk().space.DetachDataPartition(dp.partitionID)
			dp.Disk().DetachDataPartition(dp)
			dp.Stop()
			log.LogCriticalf("partition [id: %v, disk: %v] apply member change [index: %v] failed and stopped: %v",
				dp.partitionID, dp.Disk().Path, index, err)
			return
		}
		dp.advanceApplyID(index)
		log.LogWarnf("partition [%v] apply member change [index: %v] %v %v", dp.partitionID, index, confChange.Type, confChange.Peer)
	}(index)

	// Change memory the status
	var (
		isUpdated bool
	)

	switch confChange.Type {
	case raftproto.ConfAddNode:
		req := &proto.AddDataPartitionRaftMemberRequest{}
		if err = json.Unmarshal(confChange.Context, req); err != nil {
			return
		}
		isUpdated, err = dp.addRaftNode(req, index)
	case raftproto.ConfRemoveNode:
		req := &proto.RemoveDataPartitionRaftMemberRequest{}
		if err = json.Unmarshal(confChange.Context, req); err != nil {
			return
		}
		isUpdated, err = dp.removeRaftNode(req, index)
	case raftproto.ConfAddLearner:
		req := &proto.AddDataPartitionRaftLearnerRequest{}
		if err = json.Unmarshal(confChange.Context, req); err != nil {
			return
		}
		isUpdated, err = dp.addRaftLearner(req, index)
	case raftproto.ConfPromoteLearner:
		req := &proto.PromoteDataPartitionRaftLearnerRequest{}
		if err = json.Unmarshal(confChange.Context, req); err != nil {
			return
		}
		isUpdated, err = dp.promoteRaftLearner(req, index)
	case raftproto.ConfUpdateNode:
		log.LogDebugf("[updateRaftNode]: not support.")
	}
	if err != nil {
		log.LogErrorf("action[ApplyMemberChange] dp(%v) type(%v) err(%v).", dp.partitionID, confChange.Type, err)
		return
	}
	if isUpdated {
		dp.DataPartitionCreateType = proto.NormalCreateDataPartition
		if err = dp.persist(nil); err != nil {
			log.LogErrorf("action[ApplyMemberChange] dp(%v) PersistMetadata err(%v).", dp.partitionID, err)
			return
		}
	}
	dp.proposeUpdateVolumeInfo()
	return
}

// Snapshot persists the in-memory data (as a snapshot) to the disk.
// Note that the data in each data partition has already been saved on the disk. Therefore there is no need to take the
// snapshot in this case.
func (dp *DataPartition) handleRaftSnapshot(recoverNode uint64) (raftproto.Snapshot, error) {
	var statusSnap = dp.applyStatus.Snap()
	var snapshotIndex = statusSnap.NextTruncate()
	snapIterator := NewItemIterator(snapshotIndex)
	if log.IsInfoEnabled() {
		log.LogInfof("partition[%v] [lastTruncate: %v, nextTruncate: %v, applied: %v] generate raft snapshot [index: %v] for peer[%v]",
			dp.partitionID, statusSnap.LastTruncate(), statusSnap.NextTruncate(), statusSnap.Applied(), snapshotIndex, recoverNode)
	}
	return snapIterator, nil
}

// ApplySnapshot asks the raft leader for the snapshot data to recover the contents on the local disk.
func (dp *DataPartition) handleRaftApplySnapshot(peers []raftproto.Peer, iterator raftproto.SnapIterator, snapV uint32) (err error) {
	// Never delete the raft log which hadn't applied, so snapshot no need.
	log.LogInfof("PartitionID(%v) ApplySnapshot from(%v)", dp.partitionID, dp.raftPartition.CommittedIndex())
	if dp.isCatchUp {
		msg := fmt.Sprintf("partition [id: %v, disk: %v] triggers an illegal raft snapshot recover and will be stop for data safe",
			dp.partitionID, dp.Disk().Path)
		log.LogErrorf(msg)
		dp.Disk().space.DetachDataPartition(dp.partitionID)
		dp.Disk().DetachDataPartition(dp)
		dp.Stop()
		exporter.WarningCritical(msg)
		log.LogCritical(msg)
	}
	defer func() {
		dp.isCatchUp = true
	}()
	for {
		if _, err = iterator.Next(); err != nil {
			if err != io.EOF {
				log.LogError(fmt.Sprintf("action[ApplySnapshot] PartitionID(%v) ApplySnapshot from(%v) failed,err:%v", dp.partitionID, dp.raftPartition.CommittedIndex(), err.Error()))
				return
			}
			return nil
		}
	}
}

// HandleFatalEvent notifies the application when panic happens.
func (dp *DataPartition) handleRaftFatalEvent(err *raft.FatalError) {
	dp.checkIsDiskError(err.Err)
	log.LogErrorf("action[HandleFatalEvent] err(%v).", err)
}

// HandleLeaderChange notifies the application when the raft leader has changed.
func (dp *DataPartition) handleRaftLeaderChange(leader uint64) {
	defer func() {
		if r := recover(); r != nil {
			mesg := fmt.Sprintf("HandleLeaderChange(%v)  Raft Panic(%v)", dp.partitionID, r)
			panic(mesg)
		}
	}()
	if dp.config.NodeID == leader {
		if !gHasLoadDataPartition {
			go dp.raftPartition.TryToLeader(dp.partitionID)
		}
		dp.isRaftLeader = true
	}
	//If leader changed, that indicates the raft has elected a new leader,
	//the fault occurred checking to prevent raft brain split is no more needed.
	if dp.isNeedFaultCheck() {
		dp.setNeedFaultCheck(false)
		_ = dp.persistMetaDataOnly()
	}
}

func (dp *DataPartition) handleRaftAskRollback(original []byte) (rollback []byte, err error) {
	if len(original) == 0 {
		return
	}
	var opItem *rndWrtOpItem
	if opItem, err = UnmarshalRandWriteRaftLog(original); err != nil {
		return
	}
	defer func() {
		PutRandomWriteOpItem(opItem)
	}()
	var buf = make([]byte, opItem.size)
	var crc uint32
	if crc, err = dp.extentStore.Read(opItem.extentID, opItem.offset, opItem.size, buf, false); err != nil {
		return
	}
	rollback, err = MarshalRandWriteRaftLog(opItem.opcode, opItem.extentID, opItem.offset, opItem.size, buf, crc)
	log.LogWarnf("partition [id: %v, disk: %v] handle ask rollback [extent: %v, offset: %v, size: %v], CRC[%v -> %v]",
		dp.partitionID, dp.disk.Path, opItem.extentID, opItem.offset, opItem.size, opItem.crc, crc)
	return
}

// Put submits the raft log to the raft store.
func (dp *DataPartition) Put(ctx context.Context, key interface{}, val interface{}) (resp interface{}, err error) {
	if dp.raftPartition == nil {
		err = fmt.Errorf("%s key=%v", RaftNotStarted, key)
		return
	}
	resp, err = dp.raftPartition.SubmitWithCtx(ctx, val.([]byte))
	return
}

// Get returns the raft log based on the given key. It is not needed for replicating data partition.
func (dp *DataPartition) Get(key interface{}) (interface{}, error) {
	return nil, nil
}

// Del deletes the raft log based on the given key. It is not needed for replicating data partition.
func (dp *DataPartition) Del(key interface{}) (interface{}, error) {
	return nil, nil
}

func (dp *DataPartition) advanceApplyID(applyID uint64) {
	if snap, success := dp.applyStatus.AdvanceApplied(applyID); !success {
		log.LogWarnf("Partition(%v) advance apply ID failed, curApplied[%v] curLastTruncate[%v]", dp.partitionID, snap.Applied(), snap.LastTruncate())
	}
	if !dp.isCatchUp {
		dp.isCatchUp = true
	}
}
