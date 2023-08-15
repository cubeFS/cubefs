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
	"context"
	"encoding/json"
	"fmt"
	"github.com/cubefs/cubefs/proto"
	"github.com/cubefs/cubefs/util/errors"
	"github.com/cubefs/cubefs/util/exporter"
	"github.com/cubefs/cubefs/util/log"
	"math"
	"sync/atomic"
	"time"
)

func (mp *metaPartition) RecoverDeletedDentry(req *RecoverDeletedDentryReq, p *Packet) (err error) {
	entry := newPrimaryDeletedDentry(req.ParentID, req.Name, req.TimeStamp, req.Inode)
	var val []byte
	val, err = entry.Marshal()
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	var ret interface{}
	ret, err = mp.submit(p.Ctx(), opFSMRecoverDeletedDentry, p.RemoteWithReqID(), val, nil)
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return
	}
	status := ret.(*fsmOpDeletedDentryResponse).Status
	msg := ret.(*fsmOpDeletedDentryResponse).Msg
	var reply []byte
	if status != proto.OpOk {
		p.PacketErrorWithBody(status, reply)
		return
	}

	resp := new(proto.RecoverDeletedDentryResponse)
	resp.Inode = msg.Inode
	resp.Name = msg.Name
	reply, err = json.Marshal(resp)
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	p.PacketOkWithBody(reply)
	return
}

func (mp *metaPartition) BatchRecoverDeletedDentry(req *BatchRecoverDeletedDentryReq, p *Packet) (err error) {
	dens := make(DeletedDentryBatch, 0, len(req.Dens))
	for _, den := range req.Dens {
		entry := newPrimaryDeletedDentry(den.ParentID, den.Name, den.Timestamp, den.Inode)
		dens = append(dens, entry)
	}
	var val []byte
	val, err = dens.Marshal()
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	var ret interface{}
	ret, err = mp.submit(p.Ctx(), opFSMBatchRecoverDeletedDentry, p.RemoteWithReqID(), val, nil)
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return
	}
	var reply []byte
	resp := buildProtoBatchOpDeletedResp(ret.([]*fsmOpDeletedDentryResponse))
	reply, err = json.Marshal(resp)
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	p.PacketOkWithBody(reply)
	return
}

func (mp *metaPartition) BatchCleanDeletedDentry(req *BatchCleanDeletedDentryReq, p *Packet) (err error) {
	dens := make(DeletedDentryBatch, 0, len(req.Dens))
	for _, den := range req.Dens {
		entry := newPrimaryDeletedDentry(den.ParentID, den.Name, den.Timestamp, den.Inode)
		dens = append(dens, entry)
	}
	var val []byte
	val, err = dens.Marshal()
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	var ret interface{}
	ret, err = mp.submit(p.Ctx(), opFSMBatchCleanDeletedDentry, p.RemoteWithReqID(), val, nil)
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return
	}
	var reply []byte
	resp := buildProtoBatchOpDeletedResp(ret.([]*fsmOpDeletedDentryResponse))
	reply, err = json.Marshal(resp)
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	p.PacketOkWithBody(reply)
	return
}

func (mp *metaPartition) CleanDeletedDentry(req *CleanDeletedDentryReq, p *Packet) (err error) {
	entry := newPrimaryDeletedDentry(req.ParentID, req.Name, req.Timestamp, req.Inode)
	var val []byte
	val, err = entry.Marshal()
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	var ret interface{}
	ret, err = mp.submit(p.Ctx(), opFSMCleanDeletedDentry, p.RemoteWithReqID(), val, nil)
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return
	}
	var reply []byte
	status := ret.(*fsmOpDeletedDentryResponse).Status
	if status != proto.OpOk {
		p.PacketErrorWithBody(status, reply)
		return
	}

	p.PacketOkWithBody(reply)
	return
}

func (mp *metaPartition) LookupDeleted(req *LookupDeletedDentryReq, p *Packet) (err error) {
	start := newPrimaryDeletedDentry(req.ParentID, req.Name, req.StartTime, 0)
	end := newPrimaryDeletedDentry(req.ParentID, req.Name, req.EndTime, 0)
	var (
		ddentry []*DeletedDentry
		status  uint8
		reply   []byte
	)
	ddentry, status, err = mp.getDeletedDentry(start, end)
	if err != nil {
		p.PacketErrorWithBody(status, []byte(err.Error()))
		return
	}
	if status != proto.OpOk {
		p.PacketErrorWithBody(status, reply)
		return
	}
	resp := new(proto.LookupDeletedDentryResponse)
	resp.Dentrys = make([]*proto.DeletedDentry, len(ddentry))
	for index, d := range ddentry {
		pd := new(proto.DeletedDentry)
		pd.ParentID = d.ParentId
		pd.Name = d.Name
		pd.Inode = d.Inode
		pd.Type = d.Type
		pd.Timestamp = d.Timestamp
		pd.From = d.From
		resp.Dentrys[index] = pd
	}
	reply, err = json.Marshal(resp)
	if err != nil {
		status = proto.OpErr
		reply = []byte(err.Error())
		p.PacketErrorWithBody(status, reply)
		return
	}
	p.PacketOkWithBody(reply)
	return
}

func (mp *metaPartition) getDeletedDentry(start, end *DeletedDentry) (res []*DeletedDentry, status uint8, err error) {
	status = proto.OpOk
	res = make([]*DeletedDentry, 0)
	defer func() {
		log.LogDebugf("[getDeletedDentry], start: %v, end: %v, count: %v, status: %v",
			start, end, len(res), status)
		if err == rocksDBError {
			exporter.WarningRocksdbError(fmt.Sprintf("action[getDeletedDentry] clusterID[%s] volumeName[%s] partitionID[%v]"+
				" get deleted dentry failed witch rocksdb error[deleted dentry start:%v, end:%v]", mp.manager.metaNode.clusterId, mp.config.VolName,
				mp.config.PartitionId, start, end))
		}
	}()

	var dd *DeletedDentry
	if start.Timestamp == end.Timestamp && start.Timestamp > 0 {
		if dd, err = mp.dentryDeletedTree.RefGet(start.ParentId, start.Name, start.Timestamp); err != nil {
			log.LogErrorf("[getDeletedDentry] failed to get delDentry, delDentry:%v, err:%v", start, err)
			status = proto.OpErr
			return
		}
		if dd != nil {
			res = append(res, dd)
		}
	} else {
		prefix := newPrimaryDeletedDentry(start.ParentId, start.Name, 0, 0)
		err = mp.dentryDeletedTree.RangeWithPrefix(prefix, start, end, func(dd *DeletedDentry) (bool, error) {
			res = append(res, dd)
			return true, nil
		})
		if err != nil {
			log.LogErrorf("[getDeletedDentry] failed to range delDentry tree, err:%v", err)
			status = proto.OpErr
			return
		}
	}
	if len(res) == 0 {
		status = proto.OpNotExistErr
	}
	return
}

func (mp *metaPartition) ReadDeletedDir(req *ReadDeletedDirReq, p *Packet) (err error) {
	var resp *ReadDeletedDirResp
	resp, err = mp.readDeletedDir(req)
	if err != nil {
		log.LogErrorf("[ReadDeletedDir], failed to read deleted directory, err:%v", err)
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	var reply []byte
	reply, err = json.Marshal(resp)
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	p.PacketOkWithBody(reply)
	return
}

func (mp *metaPartition) readDeletedDir(req *ReadDeletedDirReq) (resp *ReadDeletedDirResp, err error) {
	resp = new(ReadDeletedDirResp)
	startDentry := newPrimaryDeletedDentry(req.ParentID, req.Name, req.Timestamp, 0)
	endDentry := newPrimaryDeletedDentry(req.ParentID+1, "", 0, 0)
	batchNum := proto.ReadDeletedDirBatchNum
	if req.Timestamp == 0 {
		batchNum = proto.ReadDeletedDirBatchNum - 1
	}
	count := 0
	prefix := newPrimaryDeletedDentry(req.ParentID, "", 0, 0)
	err = mp.dentryDeletedTree.RangeWithPrefix(prefix, startDentry, endDentry, func(dd *DeletedDentry) (bool, error) {
		count++
		// discard the first record
		if req.Timestamp > 0 && count == 1 {
			return true, nil
		}
		resp.Children = append(resp.Children, buildProtoDeletedDentry(dd))

		if count > batchNum {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		log.LogErrorf("[readDeletedDir], failed to range deletedDentry tree, error:%v", err)
		err = errors.NewErrorf("failed to range delDentry tree:%v", err)
		return
	}
	return
}

func (mp *metaPartition) getCleanTrashItemMaxDurationEachTime() (cleanDurationEachTime int32){
	cleanDurationEachTime = defCleanTrashItemMaxDurationEachTime
	//cluster config
	if val := atomic.LoadInt32(&nodeInfo.CleanTrashItemMaxDurationEachTime); val != 0 {
		cleanDurationEachTime = val
	}
	//volume config
	if val := mp.manager.getCleanTrashItemMaxDurationEachTime(mp.config.VolName); val != 0 {
		cleanDurationEachTime = val
	}
	return
}

func (mp *metaPartition) getCleanTrashItemMaxCountEachTime() (cleanMaxCountEachTime int32) {
	cleanMaxCountEachTime = defCleanTrashItemMaxTotalCountEachTime
	if val := atomic.LoadInt32(&nodeInfo.CleanTrashItemMaxCountEachTime); val != 0 {
		cleanMaxCountEachTime = val
	}
	if val := mp.manager.getCleanTrashItemMaxCountEachTime(mp.config.VolName); val != 0 {
		cleanMaxCountEachTime = val
	}
	return
}

//todo:add test
func (mp *metaPartition) CleanExpiredDeletedDentry() (err error) {
	var (
		cleanMaxCountEachTime = mp.getCleanTrashItemMaxCountEachTime()
		cleanMaxDurationEachTime = mp.getCleanTrashItemMaxDurationEachTime()
	)
	ctx := context.Background()
	fsmFunc := func(dens DeletedDentryBatch) (err error) {
		log.LogDebugf("[CleanExpiredDeletedDentry], vol:%v, mp:%v, deletedDentryCnt:%v", mp.config.VolName, mp.config.PartitionId, len(dens))
		var data []byte
		data, err = dens.Marshal()
		if err != nil {
			log.LogError(err.Error())
			return
		}
		_, err = mp.submit(ctx, opFSMCleanExpiredDentry, mp.manager.metaNode.localAddr, data, nil)
		if err != nil {
			log.LogErrorf("[CleanExpiredDeletedDentry], vol: %v, err: %v", mp.config.VolName, err.Error())
		}
		return
	}

	if mp.config.TrashRemainingDays < 0 {
		err = fmt.Errorf("[CleanExpiredDeletedDentry], vol: %v, pid: %v, trashDays: %v is invalid",
			mp.config.VolName, mp.config.PartitionId, mp.config.TrashRemainingDays)
		return
	}

	var expires int64 = math.MaxInt64
	if mp.config.TrashRemainingDays > 0 {
		expires = time.Now().AddDate(0, 0, 0-int(mp.config.TrashRemainingDays)).UnixNano() / 1000
	}

	var total int32 = 0
	defer log.LogInfof("[CleanExpiredDeletedDentry], vol: %v, mp: %v, cleaned %v until %v", mp.config.VolName, mp.config.PartitionId, total, expires)
	batch := int(mp.GetBatchDelInodeCnt() * 2)
	dens := make(DeletedDentryBatch, 0, batch)
	snap := mp.GetSnapShot()
	if snap == nil {
		err = fmt.Errorf("[CleanExpiredDeletedDentry] mp(%v) tree snap is nil", mp.config.PartitionId)
		return
	}
	defer snap.Close()
	startTime := time.Now()
	err = snap.Range(DelDentryType, func(item interface{}) (bool, error) {
		if total > cleanMaxCountEachTime || time.Since(startTime) > time.Minute * time.Duration(cleanMaxDurationEachTime) {
			log.LogInfof("[CleanExpiredDeletedDentry] mp(%v) clean Count:%v, clean time:%v",
				mp.config.PartitionId, total, time.Since(startTime).Seconds())
			return false, nil
		}
		dd := item.(*DeletedDentry)
		_, ok := mp.IsLeader()
		if !ok {
			return false, nil
		}
		if dd.Timestamp >= expires {
			return true, nil
		}
		dens = append(dens, dd)
		if len(dens) < batch {
			return true, nil
		}

		err = fsmFunc(dens)
		if err != nil {
			log.LogErrorf("[CleanExpiredDeletedDentry], vol: %v, mp: %v, err: %v", mp.config.VolName, mp.config.PartitionId, err.Error())
			return false, err
		}
		total += int32(batch)
		dens = make(DeletedDentryBatch, 0, batch)
		if mp.config.TrashRemainingDays > 0 {
			expires = time.Now().AddDate(0, 0, 0-int(mp.config.TrashRemainingDays)).UnixNano() / 1000
		} else {
			expires = math.MaxInt64
		}
		//todo:test whether needed remove
		time.Sleep(1 * time.Second)
		return true, nil
	})
	if err != nil {
		log.LogErrorf("[CleanExpiredDeletedDentry], vol: %v, mp: %v, err: %v", mp.config.VolName, mp.config.PartitionId, err.Error())
		return
	}

	_, ok := mp.IsLeader()
	if !ok {
		return
	}

	if len(dens) == 0 {
		return
	}

	err = fsmFunc(dens)
	if err != nil {
		log.LogErrorf("[CleanExpiredDeletedDentry], vol: %v, mp: %v, err: %v", mp.config.VolName, mp.config.PartitionId, err.Error())
		return
	}
	total += int32(len(dens))
	return
}

func buildProtoDeletedDentry(dd *DeletedDentry) *proto.DeletedDentry {
	var entry proto.DeletedDentry
	entry.ParentID = dd.ParentId
	entry.Name = dd.Name
	entry.Inode = dd.Inode
	entry.Type = dd.Type
	entry.Timestamp = dd.Timestamp
	entry.From = dd.From
	return &entry
}

func buildProtoOpDeletedResp(den *DeletedDentry, status uint8) (ret *proto.OpDeletedDentryRsp) {
	ret = new(proto.OpDeletedDentryRsp)
	ret.Status = status
	ret.Den = buildProtoDeletedDentry(den)
	return
}

func buildProtoBatchOpDeletedResp(resp []*fsmOpDeletedDentryResponse) (ret *proto.BatchOpDeletedDentryRsp) {
	ret = new(proto.BatchOpDeletedDentryRsp)
	if resp == nil {
		ret.Dens = make([]*proto.OpDeletedDentryRsp, 0)
		return
	}
	ret.Dens = make([]*proto.OpDeletedDentryRsp, 0, len(resp))
	for _, r := range resp {
		ret.Dens = append(ret.Dens, buildProtoOpDeletedResp(r.Msg, r.Status))
	}
	return
}
