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
	"time"

	"github.com/cubefs/cubefs/proto"
)

// Network protocol
const (
	DataNodeLatestVersion = proto.BaseVersion
)

const (
	IntervalToUpdateReplica       = 600 // interval to update the replica
	IntervalToUpdatePartitionSize = 60  // interval to update the partition size
	NumOfFilesToRecoverInParallel = 8   // number of files to be recovered simultaneously
	RaftLogRecoverInAdvance       = 5   // Partition启动时回放日志提前量
)

// Network protocol
const (
	NetworkProtocol = "tcp"
)

// Status of load data partition extent header
const (
	FinishLoadDataPartitionExtentHeader = 1
)

// cmd response
const (
	ReadFlag  = 1
	WriteFlag = 2
)

// Error code
const (
	RaftNotStarted = "RaftNotStarted"
)

// Action description
const (
	ActionNotifyFollowerToRepair          = "ActionNotifyFollowerRepair"
	ActionStreamRead                      = "ActionStreamRead"
	ActionCreateExtent                    = "ActionCreateExtent:"
	ActionMarkDelete                      = "ActionMarkDelete:"
	ActionBatchDeleteExtent               = "ActionBatchDeleteExtent"
	ActionGetAllExtentWatermarks          = "ActionGetAllExtentWatermarks:"
	ActionGetAllExtentWatermarksV2        = "ActionGetAllExtentWatermarksV2:"
	ActionGetAllExtentInfo                = "ActionGetAllExtentInfo:"
	ActionWrite                           = "ActionWrite:"
	ActionRepair                          = "ActionRepair:"
	ActionDecommissionPartition           = "ActionDecommissionPartition"
	ActionAddDataPartitionRaftMember      = "ActionAddDataPartitionRaftMember"
	ActionRemoveDataPartitionRaftMember   = "ActionRemoveDataPartitionRaftMember"
	ActionAddDataPartitionRaftLearner     = "ActionAddDataPartitionRaftLearner"
	ActionPromoteDataPartitionRaftLearner = "ActionPromoteDataPartitionRaftLearner"
	ActionDataPartitionTryToLeader        = "ActionDataPartitionTryToLeader"
	ActionResetDataPartitionRaftMember    = "ActionResetDataPartitionRaftMember"

	ActionCreateDataPartition        = "ActionCreateDataPartition"
	ActionLoadDataPartition          = "ActionLoadDataPartition"
	ActionDeleteDataPartition        = "ActionDeleteDataPartition"
	ActionStreamReadTinyDeleteRecord = "ActionStreamReadTinyDeleteRecord"
	ActionSyncTinyDeleteRecord       = "ActionSyncTinyDeleteRecord"
	ActionStreamReadTinyExtentRepair = "ActionStreamReadTinyExtentRepair"
	ActionStreamReadTinyExtentAvali  = "ActionStreamReadTinyExtentAvali"
	ActionSyncDataPartitionReplicas  = "ActionSyncDataPartitionReplicas"
	ActionGetPartitionSize           = "ActionGetPartitionSize"
	GetMaxExtentIDAndPartitionSize   = "GetMaxExtentIDAndPartitionSize"
)

// Apply the raft log operation. Currently we only have the random write operation.
const (
	MinTinyExtentsToRepair = 10 // minimum number of tiny extents to repair
)

// Tiny extent has been put back to store
const (
	IsReleased = 1
)

const (
	MinAvaliTinyExtentCnt = 5
)

// Sector size
const (
	DiskSectorSize = 512
)

const (
	RepairRead = true
	StreamRead = false
)

const (
	BufferWrite = false
)

const (
	EmptyResponse                      = 'E'
	TinyExtentRepairReadResponseArgLen = 17
	MaxSyncTinyDeleteBufferSize        = 2400000
	MaxFullSyncTinyDeleteTime          = 3600 * 24 * 8
	MinSyncTinyDeleteTime              = 3600
	MinTinyExtentDeleteRecordSyncSize  = 4 * 1024 * 1024
	DiskMaxFDLimit                     = 20000
	DiskForceEvictFDRatio              = 0.25
	CacheCapacityPerPartition          = 256
	DiskLoadPartitionParallelism       = 10
)

// data partition validate crc
const (
	EmptyCrcValue uint64 = 4045511210

	DefaultIntervalDataPartitionValidateCRC     = time.Hour * 24
	GetRemoteExtentInfoForValidateCRCRetryTimes = 3
)

const (
	errorConnRefusedMsg      = "connection refused"
	errorIOTimeoutMsg        = "i/o timeout"
	errorGetConnectMsg       = "get connection failed"
	errorPartitionRecoverMsg = "partition is recovering"
)
