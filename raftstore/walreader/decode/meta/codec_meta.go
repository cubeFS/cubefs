package meta

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/cubefs/cubefs/metanode"
	"github.com/cubefs/cubefs/proto"
	"github.com/cubefs/cubefs/raftstore/walreader/common"
	se "github.com/cubefs/cubefs/util/sortedextent"
)

const (
	metadataOpFSMCreateInode uint32 = iota
	metadataOpFSMUnlinkInode
	metadataOpFSMCreateDentry
	metadataOpFSMDeleteDentry
	metadataOpFSMDeletePartition
	metadataOpFSMUpdatePartition
	metadataOpFSMDecommissionPartition
	metadataOpFSMExtentsAdd
	metadataOpFSMStoreTick
	metadataOpStartStoreTick
	metadataOpStopStoreTick
	metadataOpFSMUpdateDentry
	metadataOpFSMExtentTruncate
	metadataOpFSMCreateLinkInode
	metadataOpFSMEvictInode
	metadataOpFSMInternalDeleteInode
	metadataOpFSMSetAttr
	metadataOpFSMInternalDelExtentFile
	metadataOpFSMInternalDelExtentCursor
	metadataOpExtentFileSnapshot
	metadataOpFSMSetXAttr
	metadataOpFSMRemoveXAttr
	metadataOpFSMCreateMultipart
	metadataOpFSMRemoveMultipart
	metadataOpFSMAppendMultipart
	metadataOpFSMSyncCursor
	//supplement action
	metadataOpFSMInternalDeleteInodeBatch
	metadataOpFSMDeleteDentryBatch
	metadataOpFSMUnlinkInodeBatch
	metadataOpFSMEvictInodeBatch
	metadataOpFSMCursorReset
	metadataOpFSMExtentsInsert
	metadataOpFSMBatchCreate
	metadataOpFSMSnapShotCrc

	metadataOpFSMCreateDeletedInode
	metadataOpFSMCreateDeletedDentry
	metadataOpFSMRecoverDeletedDentry
	metadataOpFSMBatchRecoverDeletedDentry
	metadataOpFSMRecoverDeletedInode
	metadataOpFSMBatchRecoverDeletedInode
	metadataOpFSMCleanDeletedDentry
	metadataOpFSMBatchCleanDeletedDentry
	metadataOpFSMCleanDeletedInode
	metadataOpFSMBatchCleanDeletedInode
	metadataOpFSMInternalCleanDeletedInode
	metadataOpFSMCleanExpiredDentry
	metadataOpFSMCleanExpiredInode
	metadataOpFSMExtentDelSync
	metadataOpSnapSyncExtent
	metadataOpFSMExtentMerge
	metadataOpFSMResetStoreTick
	metadataOpFSMExtentDelSyncV2
	metadataOpFSMMetaAddVirtualMP // Deprecated
	metadataOpFSMSynVirtualMPs // Deprecated

	metadataOpFSMSyncMetaConf
	metadataOpFSMSyncEvictReqRecords
)

const (
	DecoderName = "meta"
)

const (
	columnWidthFrom  = 15
	columnWidthTime  = 20
	columnWidthOp    = 24
	columnWidthAttrs = 0
)

type MetadataOpKvData struct {
	Op        uint32 `json:"op"`
	K         string `json:"k"`
	V         []byte `json:"v"`
	From      string `json:"frm"`
	Timestamp int64  `json:"ts"`
}

type MetadataCommandDecoder struct {
}

func (*MetadataCommandDecoder) Name() string {
	return DecoderName
}

func (*MetadataCommandDecoder) Header() common.ColumnValues {
	var values = common.NewColumnValues()
	values.Add(
		common.ColumnValue{Value: "FROM", Width: columnWidthFrom},
		common.ColumnValue{Value: "TIME", Width: columnWidthTime},
		common.ColumnValue{Value: "OP", Width: columnWidthOp},
		common.ColumnValue{Value: "ATTRIBUTES", Width: columnWidthAttrs},
	)
	return values
}

func (decoder *MetadataCommandDecoder) DecodeCommand(command []byte) (values common.ColumnValues, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			return
		}
	}()

	var opKVData = new(MetadataOpKvData)
	if err = json.Unmarshal(command, opKVData); err != nil {
		return
	}

	var (
		columnValFrom  = common.ColumnValue{Width: columnWidthFrom}
		columnValTime  = common.ColumnValue{Width: columnWidthTime}
		columnValOp    = common.ColumnValue{Width: columnWidthOp}
		columnValAttrs = common.ColumnValue{Width: columnWidthAttrs}
	)

	if opKVData.From != "" {
		columnValFrom.Value = opKVData.From
	} else {
		columnValFrom.Value = "N/A"
	}
	if opKVData.Timestamp > time.Now().Unix() {
		columnValTime.Value = time.Unix(opKVData.Timestamp/1000/1000, 0).Format("2006-01-02 15:04:05")
	} else if opKVData.Timestamp > 0 {
		columnValTime.Value = time.Unix(opKVData.Timestamp, 0).Format("2006-01-02 15:04:05")
	} else {
		columnValTime.Value = "N/A"
	}

	switch opKVData.Op {
	case metadataOpFSMCreateInode:
		ino := metanode.NewInode(0, 0)
		if err = ino.Unmarshal(context.Background(), opKVData.V); err != nil {
			return
		}
		columnValOp.SetValue("CreateInode")
		columnValAttrs.SetValue(fmt.Sprintf("inode: %v", ino.Inode))

	case metadataOpFSMUnlinkInode:
		ino := metanode.NewInode(0, 0)
		if err = ino.Unmarshal(context.Background(), opKVData.V); err != nil {
			return
		}
		columnValOp.SetValue("UnlinkInode")
		columnValAttrs.SetValue(fmt.Sprintf("inode: %v", ino.Inode))
	case metadataOpFSMCreateDentry:
		den := &metanode.Dentry{}
		if err = den.Unmarshal(opKVData.V); err != nil {
			return
		}
		columnValOp.SetValue("CreateDentry")
		columnValAttrs.SetValue(fmt.Sprintf("parent: %v, name: %v, inode: %v, type: %v", den.ParentId, den.Name, den.Inode, den.Type))
	case metadataOpFSMDeleteDentry:
		den := &metanode.Dentry{}
		if err = den.Unmarshal(opKVData.V); err != nil {
			return
		}
		columnValOp.SetValue("DeleteDentry")
		columnValAttrs.SetValue(fmt.Sprintf("parent: %v, name: %v", den.ParentId, den.Name))
	case metadataOpFSMDeleteDentryBatch:
		db, err := metanode.DentryBatchUnmarshal(opKVData.V)
		if err != nil {
			return nil, err
		}
		var str []string
		for _, den := range db {
			str = append(str, fmt.Sprintf("parent: %v, name: %v", den.ParentId, den.Name))
		}
		columnValOp.SetValue("DeleteDentryBatch")
		columnValAttrs.SetValue(strings.Join(str, ", "))
	case metadataOpFSMExtentTruncate:
		ino := metanode.NewInode(0, 0)
		if err = ino.Unmarshal(context.Background(), opKVData.V); err != nil {
			return
		}
		columnValOp.SetValue("ExtentTruncate")
		columnValAttrs.SetValue(fmt.Sprintf("inode: %v, size: %v", ino.Inode, ino.Size))
	case metadataOpFSMCreateLinkInode:
		ino := metanode.NewInode(0, 0)
		if err = ino.Unmarshal(context.Background(), opKVData.V); err != nil {
			return
		}
		columnValOp.SetValue("LinkInode")
		columnValAttrs.SetValue(fmt.Sprintf("inode: %v", ino.Inode))
	case metadataOpFSMEvictInode:
		ino := metanode.NewInode(0, 0)
		if err = ino.Unmarshal(context.Background(), opKVData.V); err != nil {
			return
		}
		columnValOp.SetValue("EvictInode")
		columnValAttrs.SetValue(fmt.Sprintf("inode: %v", ino.Inode))
	case metadataOpFSMInternalDeleteInode:
		buf := bytes.NewBuffer(opKVData.V)
		sb := strings.Builder{}
		ino := metanode.NewInode(0, 0)
		for {
			err = binary.Read(buf, binary.BigEndian, &ino.Inode)
			if err != nil {
				if err == io.EOF {
					err = nil
					break
				}
				return
			}
			if sb.Len() > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%v", ino.Inode))
		}
		columnValOp.SetValue("InternalDeleteInode")
		columnValAttrs.SetValue(fmt.Sprintf("inodes: %v", sb.String()))
	case metadataOpFSMSetAttr:
		req := &metanode.SetattrRequest{}
		err = json.Unmarshal(opKVData.V, req)
		if err != nil {
			return
		}
		columnValOp.SetValue("SetAttr")
		columnValAttrs.SetValue(fmt.Sprintf("inode: %v, gid: %v, uid: %v, mode: %v", req.Inode, req.Gid, req.Uid, req.Mode))
	case metadataOpFSMCursorReset:
		req := &proto.CursorResetRequest{}
		if err = json.Unmarshal(opKVData.V, req); err != nil {
			return
		}
		columnValOp.SetValue("CursorReset")
		columnValAttrs.SetValue(fmt.Sprintf("PartitionId: %v, resetType: %v, NewCursor: %v, Cursor: %v, Force: %v", req.PartitionId, req.CursorResetType, req.NewCursor, req.Cursor, req.Force))
	case metadataOpFSMExtentsAdd:
		ino := metanode.NewInode(0, 0)
		if err = ino.Unmarshal(context.Background(), opKVData.V); err != nil {
			return
		}
		columnValOp.SetValue("ExtentsAdd")
		columnValAttrs.SetValue(fmt.Sprintf("inode: %v, eks: %v", ino.Inode, decoder.formatExtentKeys(ino.Extents)))
	case metadataOpFSMExtentsInsert:
		ino := metanode.NewInode(0, 0)
		if err = ino.Unmarshal(context.Background(), opKVData.V); err != nil {
			return
		}
		columnValOp.SetValue("ExtentsInsert")
		columnValAttrs.SetValue(fmt.Sprintf("inode: %v, eks: %v", ino.Inode, decoder.formatExtentKeys(ino.Extents)))
	case metadataOpFSMSyncCursor:
		var cursor uint64
		cursor = binary.BigEndian.Uint64(opKVData.V)
		columnValOp.SetValue("SyncCursor")
		columnValAttrs.SetValue(fmt.Sprintf("cursor: %v", cursor))
	case metadataOpFSMStoreTick:
		columnValOp.SetValue("StoreTick")
	case metadataOpStartStoreTick:
		columnValOp.SetValue("StartStoreTick")
	case metadataOpStopStoreTick:
		columnValOp.SetValue("StopStoreTick")
	case metadataOpFSMUpdateDentry:
		den := &metanode.Dentry{}
		if err = den.Unmarshal(opKVData.V); err != nil {
			return
		}
		columnValOp.SetValue("UpdateDentry")
		columnValAttrs.SetValue(fmt.Sprintf("parent: %v, name: %v, inode: %v, type: %v", den.ParentId, den.Name, den.Inode, den.Type))
	case metadataOpFSMInternalDelExtentFile:
		columnValOp.SetValue("DelExtentFile")
		columnValAttrs.SetValue(fmt.Sprintf("name: %v", string(opKVData.V)))
	case metadataOpFSMInternalDelExtentCursor:
		columnValOp.SetValue("DelExtentCursor")
		columnValAttrs.SetValue(fmt.Sprintf("cursor: %v", string(opKVData.V)))
	case metadataOpFSMInternalCleanDeletedInode:
		columnValOp.SetValue("InternalCleanDeletedInode")
		inodes := make([]uint64, 0)
		if len(opKVData.V) == 0 {
			break
		}
		buf := bytes.NewBuffer(opKVData.V)
		for {
			var inodeID uint64
			err = binary.Read(buf, binary.BigEndian, &inodeID)
			if err != nil {
				if err == io.EOF {
					err = nil
					break
				}
				break
			}
			inodes = append(inodes, inodeID)
		}
		columnValAttrs.SetValue(fmt.Sprintf("inodes:%v", inodes))
	case metadataOpFSMExtentMerge:
		var inodeMerge *metanode.InodeMerge
		inodeMerge, err = metanode.InodeMergeUnmarshal(opKVData.V)
		if err != nil {
			return
		}
		columnValOp.SetValue("ExtentMerge")
		columnValAttrs.SetValue(fmt.Sprintf("inode: %v, eks:%v", inodeMerge.Inode, decoder.formatEks(inodeMerge.NewExtents, inodeMerge.OldExtents)))
	case metadataOpFSMCleanExpiredInode:
		var batch metanode.FSMDeletedINodeBatch
		if batch, err = metanode.FSMDeletedINodeBatchUnmarshal(opKVData.V); err != nil {
			return
		}
		columnValOp.SetValue("CleanExpiredInodes")
		inodes := make([]uint64, 0, len(batch))
		for _, deletedInode := range batch {
			inodeData, _ := deletedInode.Marshal()
			inodes = append(inodes, binary.BigEndian.Uint64(inodeData))
		}
		columnValAttrs.SetValue(fmt.Sprintf("inodes:%v", inodes))
	case metadataOpFSMSyncEvictReqRecords:
		columnValOp.SetValue("SyncEvictTimestamp")
		columnValAttrs.SetValue(fmt.Sprintf("EvictTimestamp:%v", int64(binary.BigEndian.Uint64(opKVData.V))))
	default:
		columnValOp.SetValue(strconv.Itoa(int(opKVData.Op)))
		columnValAttrs.SetValue("N/A")
	}
	values = common.NewColumnValues()
	values.Add(columnValFrom, columnValTime, columnValOp, columnValAttrs)
	return
}

func (decoder *MetadataCommandDecoder) formatExtentKeys(extents *se.SortedExtents) string {
	sb := strings.Builder{}
	extents.Range(func(ek proto.ExtentKey) bool {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v_%v_%v_%v_%v", ek.FileOffset, ek.PartitionId, ek.ExtentId, ek.ExtentOffset, ek.Size))
		return true
	})
	return sb.String()
}

func (decoder *MetadataCommandDecoder) formatEks(newEks, oldEks []proto.ExtentKey) string {
	sb := strings.Builder{}
	for _, ek := range newEks {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v_%v_%v_%v_%v", ek.FileOffset, ek.PartitionId, ek.ExtentId, ek.ExtentOffset, ek.Size))
	}
	sb.WriteString("||")
	for _, ek := range oldEks {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v_%v_%v_%v_%v", ek.FileOffset, ek.PartitionId, ek.ExtentId, ek.ExtentOffset, ek.Size))
	}
	return sb.String()
}
