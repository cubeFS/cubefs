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

package metanode

import (
	"fmt"

	"net"
	"strings"
	"time"

	"github.com/cubefs/cubefs/proto"
	"github.com/cubefs/cubefs/util/log"
)

const (
	ForceClosedConnect = true
	NoClosedConnect    = false
	ProxyReadTimeoutSec  = 2 // Seconds of read timout
	ProxyWriteTimeoutSec = 2
)

// The proxy is used during the leader change. When a leader of a partition changes, the proxy forwards the request to
// the new leader.
func (m *metadataManager) serveProxy(conn net.Conn, mp MetaPartition,
	p *Packet, req interface{}) (ok bool) {
	var (
		mConn           *net.TCPConn
		leaderAddr      string
		oldLeaderAddr   string
		err             error
		reqID           = p.ReqID
		reqOp           = p.Opcode
		needTryToLeader = false
	)
	if p.IsReadMetaPkt() && p.IsFollowerReadMetaPkt() {
		log.LogDebugf("read from follower: p(%v), arg(%v)", p, p.Arg)
		return true
	}

	if leaderAddr, ok = mp.IsLeader(); ok {
		if p.IsReadMetaPkt() && mp.IsRaftHang(){
			//disk error, do nothing
			time.Sleep(time.Second * RaftHangTimeOut)
			err = fmt.Errorf("mp[%d] leader raft disk is fault, try anthoer host again", mp.GetBaseConfig().PartitionId)
			p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
			goto end
		}
		return
	}


	for retryCnt := 0; retryCnt < ProxyTryToLeaderRetryCnt; {
		needTryToLeader = false
		if leaderAddr, ok = mp.IsLeader(); ok {
			//leader changed, now is myself return
			return
		}

		if leaderAddr == "" {
			err = ErrNoLeader
			p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
			goto end
		}

		if oldLeaderAddr != leaderAddr {
			retryCnt = 0
		}
		oldLeaderAddr = leaderAddr

		mConn, err = m.connPool.GetConnect(leaderAddr)
		if err != nil {
			p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
			m.connPool.PutConnect(mConn, ForceClosedConnect)
			goto end
		}

		// send to master connection
		p.ResetPackageData(req)
		if err = p.WriteToConn(mConn, ProxyWriteTimeoutSec); err != nil {
			p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
			m.connPool.PutConnect(mConn, ForceClosedConnect)
			goto end
		}

		// read connection from the master
		if err = p.ReadFromConn(mConn, ProxyReadTimeoutSec); err != nil {
			if strings.Contains(err.Error(), "i/o timeout") {
				// leader no response, retry and set try to leader flag true
				m.connPool.PutConnect(mConn, ForceClosedConnect)
				needTryToLeader = true
				retryCnt++
				continue
			}
			p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
			m.connPool.PutConnect(mConn, ForceClosedConnect)
			goto end
		}
		break
	}

	if reqID != p.ReqID || reqOp != p.Opcode {
		log.LogErrorf("serveProxy: send and received packet mismatch: req(%v_%v) resp(%v_%v)",
			reqID, reqOp, p.ReqID, p.Opcode)
	}
	m.connPool.PutConnect(mConn, NoClosedConnect)
end:
	if needTryToLeader {
		p.PacketErrorWithBody(proto.OpErr, []byte(fmt.Sprintf("proxy to leader[%s] err:%v, try to leader", leaderAddr, err)))
	}

	leaderAddr, _ = mp.IsLeader()
	if leaderAddr == oldLeaderAddr && needTryToLeader{
		log.LogErrorf("mp[%v] leader(%s) is not response, now try to elect to be leader", mp.GetBaseConfig().PartitionId, leaderAddr)
		_ = mp.TryToLeader(mp.GetBaseConfig().PartitionId)
		errMsg := "proxy failed "
		if err != nil {
			errMsg += err.Error()
		}
		p.PacketErrorWithBody(proto.OpErr, []byte(errMsg))
	}

	m.respondToClient(conn, p)
	if err != nil {
		log.LogErrorf("[serveProxy]: req: %d - %v, %s", p.GetReqID(),
			p.GetOpMsg(), err.Error())
	}
	log.LogDebugf("[serveProxy] req: %d - %v, resp: %v", p.GetReqID(), p.GetOpMsg(),
		p.GetResultMsg())
	return
}
