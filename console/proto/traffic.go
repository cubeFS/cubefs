package proto

import (
	"github.com/cubefs/cubefs/console/model"
)

const (
	HbaseVolumeClientCountPath = "/queryJson/cfsClientCount"
	HbaseVolumeClientListPath  = "/queryJson/cfsClientList"
)

const (
	TimeFormatCompact = "20060102150405"
)

const (
	ActionDrawLineDimension int32 = iota
	DiskDrawLineDimension
)

type DataGranularity int

const (
	SecondGranularity DataGranularity = iota
	MinuteGranularity
	TenMinuteGranularity
)

const (
	IntervalTypeNull = iota
	IntervalTypeLatestTenMin
	IntervalTypeLatestOneHour
	IntervalTypeLatestOneDay
)

// 角色名称
const (
	ModuleDataNode   = "DataNode"
	ModuleMetaNode   = "MetaNode"
	ModuleObjectNode = "ObjectNode"
	ModuleFlashNode  = "FlashNode"
)

var OPMap = map[string][]string{
	ModuleDataNode: {
		"read",
		"repairRead",
		"appendWrite",
		"overWrite",
		"repairWrite",
		"markDelete",
		"batchMarkDelete",
		"flushDelete",
		"diskIOCreate",
		"diskIOWrite",
		"diskIORead",
		"diskIORemove",
		"diskIOPunch",
		"diskIOSync",
	},
	ModuleMetaNode: {
		"createInode",
		"evictInode",
		"createDentry",
		"deleteDentry",
		"lookup",
		"readDir",
		"inodeGet",
		"batchInodeGet",
		"addExtents",
		"listExtents",
		"truncate",
		"insertExtent",
		"opCreateInode",
		"opEvictInode",
		"opCreateDentry",
		"opDeleteDentry",
		"opLookup",
		"opReadDir",
		"opInodeGet",
		"opBatchInodeGet",
		"opAddExtents",
		"opListExtents",
		"opTruncate",
		"opInsertExtent",
	},
	ModuleObjectNode: {
		"HeadObject",
		"GetObject",
		"PutObject",
		"ListObjects",
		"DeleteObject",
		"CopyObject",
		"CreateMultipartUpload",
		"UploadPart",
		"CompleteMultipartUpload",
		"AbortMultipartUpload",
		"ListMultipartUploads",
		"ListParts",
	},
	ModuleFlashNode: {
		"read",
		"prepare",
		"evict",
		"hit",
		"miss",
		"expire",
	},
}

type TrafficRequest struct {
	ClusterName   string `json:"clusterName"`
	IntervalType  int    `json:"intervalType"` // 1-10min 2-30min 3-1h
	Module        string `json:"module"`
	VolumeName    string `json:"volumeName"`
	OperationType string `json:"opType"`
	TopN          int    `json:"topN"`
	OrderBy       string `json:"orderBy"`
	IpAddr        string `json:"ipAddr"`
	PageSize      int    `json:"pageSize"`
	Page          int    `json:"page"`
	StartTime     int64  `json:"startTime"` // 秒级时间戳
	EndTime       int64  `json:"endTime"`
	Zone          string `json:"zone"` //zone
	Disk          string `json:"disk"` //磁盘
}

type FlowScheduleResult struct {
	VolumeName    string `json:"volume"`
	OperationType string `json:"action"`
	Time          string `json:"time"`
	IpAddr        string `json:"ip"`
	Zone          string `json:"zone"`
	Disk          string `json:"disk"`
	PartitionID   uint64 `json:"pid"`
	Count         uint64 `json:"total_count"`
	Size          uint64 `json:"total_size"` // 单位byte
	AvgSize       uint64 `json:"avg_size"`
	Max           uint64 `json:"max_latency"`
	Avg           uint64 `json:"avg"`
	Tp99          uint64 `json:"tp99"`
}

type TrafficResponse struct {
	Total int                   `json:"total"`
	Data  []*FlowScheduleResult `json:"results"`
}

type TrafficDetailsResponse struct {
	Data [][]*FlowScheduleResult `json:"results"`
}
type TrafficLatencyResponse struct {
	Data []*FlowScheduleResult `json:"results"`
}

// 不需要时间范围 展示所有的数据 时间倒排 limit 10000
type HistoryCurveRequest struct {
	Cluster      string
	Volume       string
	ZoneName     string
	IntervalType int
	Start        int64
	End          int64
}

type ZombieVolResponse struct {
	Total int64
	Data  []*model.ConsoleVolume
}

type QueryVolOpsRequest struct {
	Cluster string
	Period  model.ZombieVolPeriod
	Action  string
	Module  string
}

type TopVolResponse struct {
	TopInode []*model.VolumeSummaryView
	TopUsed  []*model.VolumeSummaryView
}

type AbnormalVolResponse struct {
	Total            int64
	ZombieVolCount   int64
	NoDeleteVolCount int64
}