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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/cubefs/cubefs/util/log"

	"github.com/cubefs/cubefs/proto"
)

func replyInfo(info *proto.InodeInfo, ino *Inode) bool {
	ino.RLock()
	defer ino.RUnlock()
	if ino.Flag&DeleteMarkFlag > 0 {
		return false
	}
	info.Inode = ino.Inode
	info.Mode = ino.Type
	info.Size = ino.Size
	info.Nlink = ino.NLink
	info.Uid = ino.Uid
	info.Gid = ino.Gid
	info.Generation = ino.Generation
	if length := len(ino.LinkTarget); length > 0 {
		info.Target = make([]byte, length)
		copy(info.Target, ino.LinkTarget)
	}
	info.CreateTime = time.Unix(ino.CreateTime, 0)
	info.AccessTime = time.Unix(ino.AccessTime, 0)
	info.ModifyTime = time.Unix(ino.ModifyTime, 0)
	return true
}

// CreateInode returns a new inode.
func (mp *metaPartition) CreateInode(req *CreateInoReq, p *Packet) (err error) {
	var (
		inoID uint64
		val   []byte
		resp  interface{}
	)

	inoID, err = mp.genInodeID()
	if err != nil {
		p.PacketErrorWithBody(proto.OpInodeFullErr, []byte(err.Error()))
		return
	}
	ino := NewInode(inoID, req.Mode)
	ino.Uid = req.Uid
	ino.Gid = req.Gid
	ino.LinkTarget = req.Target
	val, err = ino.Marshal()
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	resp, err = mp.submit(p.Ctx(), opFSMCreateInode, p.RemoteWithReqID(), val, nil)
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return
	}
	var (
		status = resp.(uint8)
		reply  []byte
	)
	if resp.(uint8) == proto.OpOk {
		resp := &CreateInoResp{
			Info: &proto.InodeInfo{},
		}
		if replyInfo(resp.Info, ino) {
			status = proto.OpOk
			reply, err = json.Marshal(resp)
			if err != nil {
				status = proto.OpErr
				reply = []byte(err.Error())
			}
		}
	}
	p.PacketErrorWithBody(status, reply)
	return
}

// DeleteInode deletes an inode.
func (mp *metaPartition) UnlinkInode(req *UnlinkInoReq, p *Packet) (err error) {
	var (
		r   interface{}
		val []byte
		msg *InodeResponse
	)

	if _, err = mp.isInoOutOfRange(req.Inode); err != nil {
		p.PacketErrorWithBody(proto.OpInodeOutOfRange, []byte(err.Error()))
		return
	}

	ino := NewInode(req.Inode, 0)

	defer func() {
		if err != nil {
			return
		}
		status := msg.Status
		var reply []byte
		if status == proto.OpOk {
			resp := &UnlinkInoResp{
				Info: &proto.InodeInfo{},
			}
			replyInfo(resp.Info, msg.Msg)
			if reply, err = json.Marshal(resp); err != nil {
				status = proto.OpErr
				reply = []byte(err.Error())
			}
		}
		p.PacketErrorWithBody(status, reply)
	}()

	clientReq := NewRequestInfo(req.ClientID, req.ClientStartTime, p.ReqID, req.ClientIP, p.CRC, mp.removeDupClientReqEnableState())
	if previousRespCode, isDup := mp.reqRecords.IsDupReq(clientReq); isDup {
		log.LogCriticalf("UnlinkInode: dup req:%v, previousRespCode:%v", clientReq, previousRespCode)
		msg = &InodeResponse{
			Status: previousRespCode,
		}
		existIno, _ := mp.inodeTree.Get(ino.Inode)
		if existIno == nil {
			log.LogCriticalf("UnlinkInode: dup req, but inode(%v) not exist", ino.Inode)
			msg.Status = proto.OpNotExistErr
			return
		}
		msg.Msg = existIno
		return
	}

	val, err = ino.Marshal()
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	if req.NoTrash || mp.isTrashDisable() {
		r, err = mp.submit(p.Ctx(), opFSMUnlinkInode, p.RemoteWithReqID(), val, clientReq)
	} else {
		r, err = mp.submitTrash(p.Ctx(), opFSMUnlinkInode, p.RemoteWithReqID(), val, clientReq)
	}
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return
	}
	msg = r.(*InodeResponse)
	return
}

// DeleteInode deletes an inode.
func (mp *metaPartition) UnlinkInodeBatch(req *BatchUnlinkInoReq, p *Packet) (err error) {
	var (
		r     interface{}
		reply []byte
		val   []byte
	)

	if len(req.Inodes) == 0 {
		return nil
	}

	var inodes InodeBatch

	for _, id := range req.Inodes {
		inodes = append(inodes, NewInode(id, 0))
	}

	val, err = inodes.Marshal(p.Ctx())
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	if req.NoTrash || mp.isTrashDisable() {
		r, err = mp.submit(p.Ctx(), opFSMUnlinkInodeBatch, p.RemoteWithReqID(), val, nil)
	} else {
		r, err = mp.submitTrash(p.Ctx(), opFSMUnlinkInodeBatch, p.RemoteWithReqID(), val, nil)
	}
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return
	}

	result := &BatchUnlinkInoResp{}
	status := proto.OpOk
	for _, ir := range r.([]*InodeResponse) {
		if ir.Status != proto.OpOk {
			status = ir.Status
			continue
		}

		info := &proto.InodeInfo{}
		replyInfo(info, ir.Msg)
		result.Items = append(result.Items, &struct {
			Info   *proto.InodeInfo `json:"info"`
			Status uint8            `json:"status"`
		}{
			Info:   info,
			Status: ir.Status,
		})
	}

	reply, err = json.Marshal(result)
	if err != nil {
		status = proto.OpErr
		reply = []byte(err.Error())
	}
	p.PacketErrorWithBody(status, reply)
	return
}

// InodeGet executes the inodeGet command from the client.
func (mp *metaPartition) InodeGet(req *InodeGetReq, p *Packet, version uint8) (err error) {

	mp.monitorData[proto.ActionMetaInodeGet].UpdateData(0)
	var (
		reply []byte
	)

	ino := NewInode(req.Inode, 0)
	var retMsg *InodeResponse
	retMsg, err = mp.getInode(ino)
	if err != nil {
		log.LogErrorf("InodeGet: get inode(Inode:%v) err:%v", req.Inode, err)
		p.PacketErrorWithBody(retMsg.Status, []byte(err.Error()))
		return
	}

	if version == proto.OpInodeGetVersion1 && retMsg.Status == proto.OpInodeOutOfRange {
		retMsg.Status = proto.OpNotExistErr
	}

	if retMsg.Status != proto.OpOk {
		p.PacketErrorWithBody(retMsg.Status, []byte("get inode err"))
		return fmt.Errorf("errCode:%d, ino:%v, mp has inodes[%v, %v]\n",
			retMsg.Status, req.Inode, mp.config.Start, mp.config.Cursor)
	}

	status := proto.OpOk
	resp := &proto.InodeGetResponse{
		Info: &proto.InodeInfo{},
	}

	if replyInfo(resp.Info, retMsg.Msg) {
		status = proto.OpOk
		reply, err = json.Marshal(resp)
		if err != nil {
			status = proto.OpErr
			reply = []byte(err.Error())
		}
	}

	p.PacketErrorWithBody(status, reply)
	return
}

// InodeGetBatch executes the inodeBatchGet command from the client.
func (mp *metaPartition) InodeGetBatch(req *InodeGetReqBatch, p *Packet) (err error) {
	var (
		ino    = NewInode(0, 0)
		data   []byte
		retMsg *InodeResponse
	)

	mp.monitorData[proto.ActionMetaBatchInodeGet].UpdateData(0)

	resp := &proto.BatchInodeGetResponse{}
	for _, inoId := range req.Inodes {
		ino.Inode = inoId
		retMsg, err = mp.getInode(ino)
		if err == nil && retMsg.Status == proto.OpOk {
			inoInfo := &proto.InodeInfo{}
			if replyInfo(inoInfo, retMsg.Msg) {
				resp.Infos = append(resp.Infos, inoInfo)
			}
		}
	}
	data, err = json.Marshal(resp)
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	p.PacketOkWithBody(data)
	return
}

// CreateInodeLink creates an inode link (e.g., soft link).
func (mp *metaPartition) CreateInodeLink(req *LinkInodeReq, p *Packet) (err error) {
	var (
		resp   interface{}
		val    []byte
		retMsg *InodeResponse
	)

	if _, err = mp.isInoOutOfRange(req.Inode); err != nil {
		p.PacketErrorWithBody(proto.OpInodeOutOfRange, []byte(err.Error()))
		return
	}

	ino := NewInode(req.Inode, 0)

	defer func() {
		if err != nil {
			return
		}
		status := proto.OpNotExistErr
		var reply []byte
		if retMsg.Status == proto.OpOk {
			r := &LinkInodeResp{
				Info: &proto.InodeInfo{},
			}
			if replyInfo(r.Info, retMsg.Msg) {
				status = proto.OpOk
				reply, err = json.Marshal(r)
				if err != nil {
					status = proto.OpErr
					reply = []byte(err.Error())
				}
			}

		}
		p.PacketErrorWithBody(status, reply)
	}()

	clientReq := NewRequestInfo(req.ClientID, req.ClientStartTime, p.ReqID, req.ClientIP, p.CRC, mp.removeDupClientReqEnableState())
	if previousRespCode, isDup := mp.reqRecords.IsDupReq(clientReq); isDup {
		log.LogCriticalf("CreateInodeLink: dup req:%v, previousRespCode:%v", clientReq, previousRespCode)
		retMsg = &InodeResponse{
			Status: previousRespCode,
		}
		existIno, _ := mp.inodeTree.Get(ino.Inode)
		if existIno == nil {
			log.LogCriticalf("CreateInodeLink: dup req, but inode(%v) not exist", ino.Inode)
			retMsg.Status = proto.OpNotExistErr
			return
		}
		retMsg.Msg = existIno
		return
	}

	val, err = ino.Marshal()
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	resp, err = mp.submit(p.Ctx(), opFSMCreateLinkInode, p.RemoteWithReqID(), val, clientReq)
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return
	}
	retMsg = resp.(*InodeResponse)
	return
}

// EvictInode evicts an inode.
func (mp *metaPartition) EvictInode(req *EvictInodeReq, p *Packet) (err error) {
	var (
		resp interface{}
		val  []byte
	)

	if _, err = mp.isInoOutOfRange(req.Inode); err != nil {
		p.PacketErrorWithBody(proto.OpInodeOutOfRange, []byte(err.Error()))
		return
	}

	ino := NewInode(req.Inode, 0)
	val, err = ino.Marshal()
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}

	if req.NoTrash || mp.isTrashDisable() {
		resp, err = mp.submit(p.Ctx(), opFSMEvictInode, p.RemoteWithReqID(), val, nil)
	} else {
		resp, err = mp.submitTrash(p.Ctx(), opFSMEvictInode, p.RemoteWithReqID(), val, nil)
	}
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return
	}
	msg := resp.(*InodeResponse)
	p.PacketErrorWithBody(msg.Status, nil)
	return
}

// EvictInode evicts an inode.
func (mp *metaPartition) EvictInodeBatch(req *BatchEvictInodeReq, p *Packet) (err error) {
	var (
		resp interface{}
	)

	if len(req.Inodes) == 0 {
		return nil
	}

	var inodes InodeBatch

	for _, id := range req.Inodes {
		inodes = append(inodes, NewInode(id, 0))
	}

	val, err := inodes.Marshal(p.Ctx())
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	if req.NoTrash || mp.isTrashDisable() {
		resp, err = mp.submit(p.Ctx(), opFSMEvictInodeBatch, p.RemoteWithReqID(), val, nil)
	} else {
		resp, err = mp.submitTrash(p.Ctx(), opFSMEvictInodeBatch, p.RemoteWithReqID(), val, nil)
	}
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return
	}

	status := proto.OpOk
	for _, m := range resp.([]*InodeResponse) {
		if m.Status != proto.OpOk {
			status = m.Status
		}
	}

	p.PacketErrorWithBody(status, nil)
	return
}

// SetAttr set the inode attributes.
func (mp *metaPartition) SetAttr(req *SetattrRequest, reqData []byte, p *Packet) (err error) {
	var (
		resp interface{}
	)

	clientReqInfo := NewRequestInfo(req.ClientID, req.ClientStartTime, p.ReqID, req.ClientIP, p.CRC, mp.removeDupClientReqEnableState())
	if previousRespCode, isDup := mp.reqRecords.IsDupReq(clientReqInfo); isDup {
		log.LogCriticalf("setAttr: dup req:%v, previousRespCode:%v", clientReqInfo, previousRespCode)
		p.PacketErrorWithBody(previousRespCode, nil)
		return
	}

	resp, err = mp.submit(p.Ctx(), opFSMSetAttr, p.RemoteWithReqID(), reqData, clientReqInfo)
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return
	}

	if (resp.(*InodeResponse)).Status != proto.OpOk {
		p.PacketErrorWithBody(resp.(*InodeResponse).Status, []byte("Apply set attr failed"))
		return
	}

	p.PacketOkReply()
	return
}

func (mp *metaPartition) DeleteInode(req *proto.DeleteInodeRequest, p *Packet) (err error) {

	if _, err = mp.isInoOutOfRange(req.Inode); err != nil {
		p.PacketErrorWithBody(proto.OpInodeOutOfRange, []byte(err.Error()))
		return
	}

	var bytes = make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, req.Inode)
	_, err = mp.submit(p.Ctx(), opFSMInternalDeleteInode, p.RemoteWithReqID(), bytes, nil)
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return
	}
	p.PacketOkReply()
	return
}

func (mp *metaPartition) CursorReset(ctx context.Context, req *proto.CursorResetRequest) error {
	status, _ := mp.calcMPStatus()
	if status == proto.Unavailable {
		log.LogInfof("mp[%v] status[%d] is unavailable[%d], can not reset cursor[%v]",
			mp.config.PartitionId, status, proto.Unavailable, mp.config.Cursor)
		return fmt.Errorf("mp[%v] status[%d] is unavailable[%d], can not reset cursor[%v]",
			mp.config.PartitionId, status, proto.Unavailable, mp.config.Cursor)
	}

	if mp.config.End == defaultMaxMetaPartitionInodeID {
		log.LogInfof("mp[%v] is max partition, not support reset cursor", mp.config.PartitionId)
		return fmt.Errorf("max partition not support reset cursor")
	}

	maxIno := mp.config.Start
	maxInode := mp.inodeTree.MaxItem()
	if maxInode != nil {
		maxIno = maxInode.Inode
	}

	switch CursorResetMode(req.CursorResetType) {
	case SubCursor:
		if status != proto.ReadOnly {
			log.LogInfof("mp[%v] status[%d] is not readonly[%d], can not reset cursor[%v]",
				mp.config.PartitionId, status, proto.ReadOnly, mp.config.Cursor)
			return fmt.Errorf("mp[%v] status[%d] is not readonly[%d], can not reset cursor[%v]",
				mp.config.PartitionId, status, proto.ReadOnly, mp.config.Cursor)
		}

		if req.NewCursor == 0 {
			req.NewCursor = maxIno + mpResetInoStep
		}

		req.Cursor = atomic.LoadUint64(&mp.config.Cursor)
		if req.NewCursor >= req.Cursor {
			return fmt.Errorf("operation mismatch, cursorResetMode(%v), newCursor(%v) oldCursor(%v)",
				CursorResetMode(req.CursorResetType), req.NewCursor, req.Cursor)
		}

		willFree := mp.config.End - req.NewCursor
		if !req.Force && willFree < mpResetInoLimited {
			log.LogInfof("mp[%v] max inode[%v] is too high, no need reset",
				mp.config.PartitionId, maxIno)
			return fmt.Errorf("mp[%v] max inode[%v] is too high, no need reset",
				mp.config.PartitionId, maxIno)
		}
	case AddCursor:
		req.NewCursor = atomic.LoadUint64(&mp.config.End)
	default:
		return fmt.Errorf("mp[%v] with error cursor reset mode[%v]", mp.config.PartitionId, req.CursorResetType)
	}

	if req.NewCursor <= maxIno || req.NewCursor > mp.config.End {
		log.LogInfof("mp[%v] req ino[%d] is out of max[%d]~end[%d]",
			mp.config.PartitionId, req.NewCursor, maxIno, mp.config.End)
		return fmt.Errorf("mp[%v] req ino[%d] is out of max[%d]~end[%d]", mp.config.PartitionId, req.NewCursor, maxIno, mp.config.End)
	}

	data, err := json.Marshal(req)
	if err != nil {
		log.LogInfof("mp[%v] reset cursor failed, json marshal failed:%v",
			mp.config.PartitionId, err.Error())
		return err
	}

	if _, ok := mp.IsLeader(); !ok {
		return fmt.Errorf("this node is not leader, can not execute this op")
	}
	resp, err := mp.submit(ctx, opFSMCursorReset, "", data, nil)
	if err != nil {
		return err
	}
	if resp.(*CursorResetResponse).Status != proto.OpOk {
		return fmt.Errorf(resp.(*CursorResetResponse).Msg)
	}
	return nil
}

func (mp *metaPartition) DeleteInodeBatch(req *proto.DeleteInodeBatchRequest, p *Packet) (err error) {

	if len(req.Inodes) == 0 {
		return nil
	}

	var inodes InodeBatch

	for _, id := range req.Inodes {
		inodes = append(inodes, NewInode(id, 0))
	}

	encoded, err := inodes.Marshal(p.Ctx())
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return
	}
	_, err = mp.submit(p.Ctx(), opFSMInternalDeleteInodeBatch, p.RemoteWithReqID(), encoded, nil)
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return
	}
	p.PacketOkReply()
	return
}

func (mp *metaPartition) GetInodeTree() InodeTree {
	return mp.inodeTree
}

func (mp *metaPartition) InodesMergeCheck(inos []uint64, limitCnt uint32, minEkLen int, minInodeSize uint64, maxEkAvgSize uint64) (resp *proto.GetCmpInodesResponse) {
	resp = &proto.GetCmpInodesResponse{}
	cnt := uint32(0)
	for i := 0; i < len(inos) && cnt < limitCnt; i++ {
		ino := inos[i]
		ok, inode := mp.hasInode(&Inode{Inode: ino})
		if !ok {
			continue
		}

		if !proto.IsRegular(inode.Type) || !inode.IsNeedCompact(minEkLen, minInodeSize, maxEkAvgSize) {
			continue
		}

		cInode := &proto.InodeInfo{}
		if !replyInfo(cInode, inode) {
			continue
		}
		resp.Inodes = append(resp.Inodes, &proto.CmpInodeInfo{Inode: cInode, Extents: inode.Extents.CopyExtents()})
		cnt++
	}
	return
}

func (mp *metaPartition) GetCompactInodeInfo(req *proto.GetCmpInodesRequest, p *Packet) (err error) {
	var data []byte

	defer func() {
		if err != nil {
			p.PacketErrorWithBody(proto.OpNotExistErr, []byte(err.Error()))
		}
	}()

	if len(req.Inodes) == 0 {
		err = fmt.Errorf("inodes not exist")
		return
	}

	resp := mp.InodesMergeCheck(req.Inodes, req.ParallelCnt, req.MinEkLen, req.MinInodeSize, req.MaxEkAvgSize)

	if data, err = json.Marshal(resp); err != nil {
		return
	}
	p.PacketErrorWithBody(proto.OpOk, data)
	return
}
