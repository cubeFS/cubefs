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

package master

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cubefs/cubefs/util/exporter"

	"github.com/cubefs/cubefs/proto"
	pb "github.com/gogo/protobuf/proto"
)

type AdminAPI struct {
	mc *MasterClient
}

func (api *AdminAPI) GetCluster() (cv *proto.ClusterView, err error) {
	var (
		buf         []byte
		contentType string
	)
	var request = newAPIRequest(http.MethodGet, proto.AdminGetCluster)
	request.addHeader(proto.AcceptFormat, proto.ProtobufType)
	if buf, contentType, err = api.mc.serveRequest(request); err != nil {
		return
	}
	if contentType == proto.ProtobufType {
		cvPb := &proto.ClusterViewPb{}
		if err = pb.Unmarshal(buf, cvPb); err != nil {
			return
		}
		cv = proto.ConvertClusterViewPb(cvPb)
	} else {
		cv = &proto.ClusterView{}
		if err = json.Unmarshal(buf, &cv); err != nil {
			return
		}
	}
	return
}

func (api *AdminAPI) GetClientConf() (cf *proto.ClientClusterConf, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.ClientConfCluster)
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	cf = &proto.ClientClusterConf{}
	if err = json.Unmarshal(buf, &cf); err != nil {
		return
	}
	return
}
func (api *AdminAPI) GetTopology() (cv *proto.TopologyView, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.GetTopologyView)
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	cv = &proto.TopologyView{}
	if err = json.Unmarshal(buf, &cv); err != nil {
		return
	}
	return
}
func (api *AdminAPI) GetClusterStat() (cs *proto.ClusterStatInfo, err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminClusterStat)
	request.addHeader("isTimeOut", "false")
	var data []byte
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	cs = &proto.ClusterStatInfo{}
	if err = json.Unmarshal(data, &cs); err != nil {
		return
	}
	return
}
func (api *AdminAPI) GetDataPartition(volName string, partitionID uint64) (partition *proto.DataPartitionInfo, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminGetDataPartition)
	request.addParam("id", strconv.Itoa(int(partitionID)))
	request.addParam("name", volName)
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	partition = &proto.DataPartitionInfo{}
	if err = json.Unmarshal(buf, &partition); err != nil {
		return
	}
	return
}

func (api *AdminAPI) DiagnoseDataPartition() (diagnosis *proto.DataPartitionDiagnosis, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminDiagnoseDataPartition)
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	diagnosis = &proto.DataPartitionDiagnosis{}
	if err = json.Unmarshal(buf, &diagnosis); err != nil {
		return
	}
	return
}

func (api *AdminAPI) ResetDataPartition(partitionID uint64) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminResetDataPartition)
	request.addParam("id", strconv.Itoa(int(partitionID)))
	request.addHeader("isTimeOut", "false")
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) ManualResetDataPartition(partitionID uint64, nodeAddrs string) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminManualResetDataPartition)
	request.addParam("id", strconv.Itoa(int(partitionID)))
	request.addParam("addr", nodeAddrs)
	request.addHeader("isTimeOut", "false")
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) FreezeDataPartition(volName string, partitionId uint64) (result string, err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminFreezeDataPartition)
	request.addParam("id", strconv.Itoa(int(partitionId)))
	request.addParam("name", volName)
	var data []byte
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return "", err
	}
	return string(data), nil
}

func (api *AdminAPI) UnfreezeDataPartition(volName string, partitionId uint64) (result string, err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminUnfreezeDataPartition)
	request.addParam("id", strconv.Itoa(int(partitionId)))
	request.addParam("name", volName)
	var data []byte
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return "", err
	}
	return string(data), nil
}

func (api *AdminAPI) ResetCorruptDataNode(nodeAddr string) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminResetCorruptDataNode)
	request.addParam("addr", nodeAddr)
	request.addHeader("isTimeOut", "false")
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) ResetRecoverDataPartition(partitionId uint64) (result string, err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminDataPartitionSetIsRecover)
	request.addParam("id", strconv.Itoa(int(partitionId)))
	request.addParam("isRecover", "false")
	var data []byte
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return "", err
	}
	return string(data), nil
}

func (api *AdminAPI) ResetRecoverMetaPartition(partitionId uint64) (result string, err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminMetaPartitionSetIsRecover)
	request.addParam("id", strconv.Itoa(int(partitionId)))
	request.addParam("isRecover", "false")
	var data []byte
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return "", err
	}
	return string(data), nil
}

func (api *AdminAPI) DiagnoseMetaPartition() (diagnosis *proto.MetaPartitionDiagnosis, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminDiagnoseMetaPartition)
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	diagnosis = &proto.MetaPartitionDiagnosis{}
	if err = json.Unmarshal(buf, &diagnosis); err != nil {
		return
	}
	return
}
func (api *AdminAPI) ResetMetaPartition(partitionID uint64) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminResetMetaPartition)
	request.addParam("id", strconv.Itoa(int(partitionID)))
	request.addHeader("isTimeOut", "false")
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}
func (api *AdminAPI) ManualResetMetaPartition(partitionID uint64, nodeAddrs string) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminManualResetMetaPartition)
	request.addParam("id", strconv.Itoa(int(partitionID)))
	request.addParam("addr", nodeAddrs)
	request.addHeader("isTimeOut", "false")
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}
func (api *AdminAPI) ResetCorruptMetaNode(nodeAddr string) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminResetCorruptMetaNode)
	request.addParam("addr", nodeAddr)
	request.addHeader("isTimeOut", "false")
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

//func (api *AdminAPI) LoadDataPartition(volName string, partitionID uint64) (err error) {
//	var request = newAPIRequest(http.MethodGet, proto.AdminLoadDataPartition)
//	request.addParam("id", strconv.Itoa(int(partitionID)))
//	request.addParam("name", volName)
//	if _, _, err = api.mc.serveRequest(request); err != nil {
//		return
//	}
//	return
//}

func (api *AdminAPI) MockCreateDataPartition(nodeAddr string, dp uint64) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminCreateDataPartition)
	request.addParam("addr", nodeAddr)
	request.addParam("id", strconv.FormatUint(dp, 10))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}
func (api *AdminAPI) MockDeleteDataReplica(nodeAddr string, dp uint64) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminDeleteDataReplica)
	request.addParam("addr", nodeAddr)
	request.addParam("id", strconv.FormatUint(dp, 10))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) CreateDataPartition(volName string, count int) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminCreateDataPartition)
	request.addParam("name", volName)
	request.addParam("count", strconv.Itoa(count))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) DecommissionDataPartition(dataPartitionID uint64, nodeAddr, destAddr string) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminDecommissionDataPartition)
	request.addParam("id", strconv.FormatUint(dataPartitionID, 10))
	request.addParam("addr", nodeAddr)
	if len(strings.TrimSpace(destAddr)) != 0 {
		request.addParam("destAddr", destAddr)
	}
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) DecommissionMetaPartition(metaPartitionID uint64, nodeAddr, destAddr string, storeMode int) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminDecommissionMetaPartition)
	request.addParam("id", strconv.FormatUint(metaPartitionID, 10))
	request.addParam("addr", nodeAddr)
	if len(strings.TrimSpace(destAddr)) != 0 {
		request.addParam("destAddr", destAddr)
	}
	if storeMode != 0 {
		request.addParam("storeMode", strconv.Itoa(storeMode))
	}
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) DeleteDataReplica(dataPartitionID uint64, nodeAddr string) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminDeleteDataReplica)
	request.addParam("id", strconv.FormatUint(dataPartitionID, 10))
	request.addParam("addr", nodeAddr)
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) AddDataReplica(dataPartitionID uint64, nodeAddr string, addReplicaType proto.AddReplicaType) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminAddDataReplica)
	request.addParam("id", strconv.FormatUint(dataPartitionID, 10))
	request.addParam("addr", nodeAddr)
	if addReplicaType != 0 {
		request.addParam("addReplicaType", strconv.Itoa(int(addReplicaType)))
	}
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) AddDataLearner(dataPartitionID uint64, nodeAddr string, autoPromote bool, threshold uint8) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminAddDataReplicaLearner)
	request.addParam("id", strconv.FormatUint(dataPartitionID, 10))
	request.addParam("addr", nodeAddr)
	request.addParam("auto", strconv.FormatBool(autoPromote))
	request.addParam("threshold", strconv.FormatUint(uint64(threshold), 10))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) PromoteDataLearner(dataPartitionID uint64, nodeAddr string) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminPromoteDataReplicaLearner)
	request.addParam("id", strconv.FormatUint(dataPartitionID, 10))
	request.addParam("addr", nodeAddr)
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) DeleteMetaReplica(metaPartitionID uint64, nodeAddr string) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminDeleteMetaReplica)
	request.addParam("id", strconv.FormatUint(metaPartitionID, 10))
	request.addParam("addr", nodeAddr)
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) AddMetaReplica(metaPartitionID uint64, nodeAddr string, addReplicaType proto.AddReplicaType,
	storeMode int) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminAddMetaReplica)
	request.addParam("id", strconv.FormatUint(metaPartitionID, 10))
	request.addParam("addr", nodeAddr)
	if addReplicaType != 0 {
		request.addParam("addReplicaType", strconv.Itoa(int(addReplicaType)))
	}
	if storeMode != 0 {
		request.addParam("storeMode", strconv.Itoa(storeMode))
	}
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) AddMetaReplicaLearner(metaPartitionID uint64, nodeAddr string, autoPromote bool, threshold uint8,
	addReplicaType proto.AddReplicaType, storeMode int) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminAddMetaReplicaLearner)
	request.addParam("id", strconv.FormatUint(metaPartitionID, 10))
	request.addParam("addr", nodeAddr)
	request.addParam("auto", strconv.FormatBool(autoPromote))
	request.addParam("threshold", strconv.FormatUint(uint64(threshold), 10))
	if addReplicaType != 0 {
		request.addParam("addReplicaType", strconv.Itoa(int(addReplicaType)))
	}
	if storeMode != 0 {
		request.addParam("storeMode", strconv.Itoa(storeMode))
	}
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) SelectMetaReplicaReplaceNodeAddr(metaPartitionID uint64, nodeAddr string, storeMode int) (info *proto.SelectMetaNodeInfo, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminSelectMetaReplicaNode)

	request.addParam("id", strconv.FormatUint(metaPartitionID, 10))
	request.addParam("addr", nodeAddr)
	if storeMode != 0 {
		request.addParam("storeMode", strconv.Itoa(storeMode))
	}

	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	info = &proto.SelectMetaNodeInfo{}
	if err = json.Unmarshal(buf, info); err != nil {
		return
	}
	return
}

func (api *AdminAPI) PromoteMetaReplicaLearner(metaPartitionID uint64, nodeAddr string) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminPromoteMetaReplicaLearner)
	request.addParam("id", strconv.FormatUint(metaPartitionID, 10))
	request.addParam("addr", nodeAddr)
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) DeleteVolume(volName, authKey string) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminDeleteVol)
	request.addParam("name", volName)
	request.addParam("authKey", authKey)
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) ForceDeleteVolume(volName, authKey string) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminForceDeleteVol)
	request.addParam("name", volName)
	request.addParam("authKey", authKey)
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) RecoverVolume(oldVolName, authKey, newVolName string) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminRecoverVol)
	request.addParam("name", oldVolName)
	request.addParam("authKey", authKey)
	request.addParam("newName", newVolName)
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) UpdateVolume(volName string, capacity uint64, replicas, mpReplicas, trashDays, storeMode int,
	followerRead, volWriteMutex, nearRead, authenticate, enableToken, autoRepair, forceROW, isSmart, enableWriteCache bool,
	authKey, zoneName, mpLayout, smartRules string, bucketPolicy, crossRegionHAType uint8,
	extentCacheExpireSec int64, compactTag string, hostDelayInterval int64, follReadHostWeight int, trashCleanInterVal uint64,
	batchDelInodeCnt, delInodeInterval uint32, umpCollectWay exporter.UMPCollectMethod, trashCleanDuration, trashCleanMaxCount int32,
	enableBitMapAllocator bool, remoteCacheBoostPath string, remoteCacheBoostEnable, remoteCacheAutoPrepare bool,
	remoteCacheTTL int64, enableRemoveDupReq bool) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminUpdateVol)
	request.addParam("name", volName)
	request.addParam("authKey", authKey)
	request.addParam("capacity", strconv.FormatUint(capacity, 10))
	request.addParam("replicaNum", strconv.Itoa(replicas))
	request.addParam("mpReplicaNum", strconv.Itoa(mpReplicas))
	request.addParam("followerRead", strconv.FormatBool(followerRead))
	request.addParam("volWriteMutex", strconv.FormatBool(volWriteMutex))
	request.addParam("nearRead", strconv.FormatBool(nearRead))
	request.addParam("forceROW", strconv.FormatBool(forceROW))
	request.addParam("writeCache", strconv.FormatBool(enableWriteCache))
	request.addParam("authenticate", strconv.FormatBool(authenticate))
	request.addParam("enableToken", strconv.FormatBool(enableToken))
	request.addParam("autoRepair", strconv.FormatBool(autoRepair))
	request.addParam("zoneName", zoneName)
	request.addParam("bucketPolicy", strconv.Itoa(int(bucketPolicy)))
	request.addParam("crossRegion", strconv.Itoa(int(crossRegionHAType)))
	request.addParam("ekExpireSec", strconv.FormatInt(extentCacheExpireSec, 10))
	request.addParam("storeMode", strconv.Itoa(storeMode))
	request.addParam("metaLayout", mpLayout)
	request.addParam("smart", strconv.FormatBool(isSmart))
	request.addParam("smartRules", smartRules)
	request.addParam("compactTag", compactTag)
	request.addParam("hostDelayInterval", strconv.Itoa(int(hostDelayInterval)))
	request.addParam("follReadHostWeight", strconv.Itoa(follReadHostWeight))
	request.addParam("batchDelInodeCnt", strconv.Itoa(int(batchDelInodeCnt)))
	request.addParam("delInodeInterval", strconv.Itoa(int(delInodeInterval)))
	request.addParam(proto.MetaTrashCleanIntervalKey, strconv.FormatUint(trashCleanInterVal, 10))
	request.addParam(proto.EnableBitMapAllocatorKey, strconv.FormatBool(enableBitMapAllocator))
	if trashDays > -1 {
		request.addParam("trashRemainingDays", strconv.Itoa(trashDays))
	}
	request.addParam("umpCollectWay", strconv.Itoa(int(umpCollectWay)))
	if trashCleanDuration >= 0 {
		request.addParam(proto.TrashCleanDurationKey, strconv.FormatInt(int64(trashCleanDuration), 10))
	}
	if trashCleanMaxCount >= 0 {
		request.addParam(proto.TrashItemCleanMaxCountKey, strconv.FormatInt(int64(trashCleanMaxCount), 10))
	}
	request.addParam("remoteCacheBoostPath", remoteCacheBoostPath)
	request.addParam("remoteCacheBoostEnable", strconv.FormatBool(remoteCacheBoostEnable))
	request.addParam("remoteCacheAutoPrepare", strconv.FormatBool(remoteCacheAutoPrepare))
	request.addParam("remoteCacheTTL", strconv.FormatInt(remoteCacheTTL, 10))
	request.addParam(proto.VolRemoveDupFlagKey, strconv.FormatBool(enableRemoveDupReq))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) ShrinkVolCapacity(volName, authKey string, capacity uint64) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminShrinkVolCapacity)
	request.addParam("name", volName)
	request.addParam("authKey", authKey)
	request.addParam("capacity", strconv.FormatUint(capacity, 10))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) SetVolumeConvertTaskState(volName, authKey string, st int) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminSetVolConvertSt)
	request.addParam("name", volName)
	request.addParam("authKey", authKey)
	request.addParam("state", strconv.Itoa(st))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) CreateVolume(volName, owner string, mpCount int, dpSize, capacity uint64, replicas, mpReplicas, trashDays, storeMode int,
	followerRead, autoRepair, volWriteMutex, forceROW, isSmart, enableWriteCache bool, zoneName, mpLayout, smartRules string,
	crossRegionHAType uint8, compactTag string, ecDataNum, ecParityNum uint8, ecEnable bool, hostDelayInterval int64,
	batchDelInodeCnt, delInodeInterval uint64, bitMapAllocatorEnable bool) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminCreateVol)
	request.addParam("name", volName)
	request.addParam("owner", owner)
	request.addParam("mpCount", strconv.Itoa(mpCount))
	request.addParam("size", strconv.FormatUint(dpSize, 10))
	request.addParam("capacity", strconv.FormatUint(capacity, 10))
	request.addParam("ecDataNum", strconv.Itoa(int(ecDataNum)))
	request.addParam("ecParityNum", strconv.Itoa(int(ecParityNum)))
	request.addParam("ecEnable", strconv.FormatBool(ecEnable))
	request.addParam("followerRead", strconv.FormatBool(followerRead))
	request.addParam("forceROW", strconv.FormatBool(forceROW))
	request.addParam("writeCache", strconv.FormatBool(enableWriteCache))
	request.addParam("crossRegion", strconv.Itoa(int(crossRegionHAType)))
	request.addParam("autoRepair", strconv.FormatBool(autoRepair))
	request.addParam("replicaNum", strconv.Itoa(replicas))
	request.addParam("mpReplicaNum", strconv.Itoa(mpReplicas))
	request.addParam("volWriteMutex", strconv.FormatBool(volWriteMutex))
	request.addParam("zoneName", zoneName)
	request.addParam("trashRemainingDays", strconv.Itoa(trashDays))
	request.addParam("storeMode", strconv.Itoa(storeMode))
	request.addParam("metaLayout", mpLayout)
	request.addParam("smart", strconv.FormatBool(isSmart))
	request.addParam("smartRules", smartRules)
	request.addParam("compactTag", compactTag)
	request.addParam("hostDelayInterval", strconv.Itoa(int(hostDelayInterval)))
	request.addHeader("isTimeOut", "false")
	request.addParam("batchDelInodeCnt", strconv.Itoa(int(batchDelInodeCnt)))
	request.addParam("delInodeInterval", strconv.Itoa(int(delInodeInterval)))
	request.addParam(proto.EnableBitMapAllocatorKey, strconv.FormatBool(bitMapAllocatorEnable))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) CreateDefaultVolume(volName, owner string) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminCreateVol)
	request.addParam("name", volName)
	request.addParam("owner", owner)
	request.addParam("capacity", "10")
	request.addHeader("isTimeOut", "false")
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) GetVolumeSimpleInfo(volName string) (vv *proto.SimpleVolView, err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminGetVol)
	request.addParam("name", volName)
	request.addParam("baseVersion", proto.BaseVersion)
	request.addHeader(proto.AcceptFormat, proto.ProtobufType)
	var (
		data        []byte
		contentType string
	)
	if data, contentType, err = api.mc.serveRequest(request); err != nil {
		return
	}

	if contentType == proto.ProtobufType {
		vvPb := &proto.SimpleVolViewPb{}
		if err = pb.Unmarshal(data, vvPb); err != nil {
			return
		}
		vv = proto.ConvertSimpleVolViewPb(vvPb)
	} else {
		vv = &proto.SimpleVolView{}
		if err = json.Unmarshal(data, &vv); err != nil {
			return
		}
	}
	return
}

func (api *AdminAPI) GetClusterInfo() (ci *proto.ClusterInfo, err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminGetIP)
	var data []byte
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	ci = &proto.ClusterInfo{}
	if err = json.Unmarshal(data, &ci); err != nil {
		return
	}
	return
}

func (api *AdminAPI) GetLimitInfo(volName string) (info *proto.LimitInfo, err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminGetLimitInfo)
	request.addParam("name", volName)
	var data []byte
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	info = &proto.LimitInfo{}
	if err = json.Unmarshal(data, &info); err != nil {
		return
	}
	return
}

func (api *AdminAPI) CreateMetaPartition(volName string, inodeStart uint64) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminCreateMetaPartition)
	request.addParam("name", volName)
	request.addParam("start", strconv.FormatUint(inodeStart, 10))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) ListVols(keywords string) (volsInfo []*proto.VolInfo, err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminListVols)
	request.addParam("keywords", keywords)
	var data []byte
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	volsInfo = make([]*proto.VolInfo, 0)
	if err = json.Unmarshal(data, &volsInfo); err != nil {
		return
	}
	return
}

func (api *AdminAPI) ListVolsByKeywordsAndSmart(keywords, smart string) (volsInfo []*proto.VolInfo, err error) {
	var request *request
	if len(smart) == 0 {
		request = newAPIRequest(http.MethodGet, proto.AdminListVols)
		request.addParam("keywords", keywords)
	} else {
		request = newAPIRequest(http.MethodGet, proto.AdminSmartVolList)
		request.addParam("keywords", keywords)
		request.addParam("smart", smart)
	}
	var data []byte
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	volsInfo = make([]*proto.VolInfo, 0)
	if err = json.Unmarshal(data, &volsInfo); err != nil {
		return
	}
	return
}

func (api *AdminAPI) IsFreezeCluster(isFreeze bool) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminClusterFreeze)
	request.addParam("enable", strconv.FormatBool(isFreeze))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) SetMetaNodeThreshold(threshold float64) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminSetMetaNodeThreshold)
	request.addParam("threshold", strconv.FormatFloat(threshold, 'f', 6, 64))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) SetMetaNodeRocksDBDiskThreshold(threshold float64) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminSetMNRocksDBDiskThreshold)
	request.addParam("threshold", strconv.FormatFloat(threshold, 'f', 6, 64))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) SetMetaNodeMemModeRocksDBDiskThreshold(threshold float64) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminSetMNMemModeRocksDBDiskThreshold)
	request.addParam("threshold", strconv.FormatFloat(threshold, 'f', 6, 64))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) SetRateLimit(info *proto.RateLimitInfo) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminSetNodeInfo)
	if info.Opcode >= 0 {
		request.addParam("opcode", strconv.FormatInt(int64(info.Opcode), 10))
	}
	if info.DataNodeFlushFDInterval >= 0 {
		request.addParam("dataNodeFlushFDInterval", strconv.FormatInt(info.DataNodeFlushFDInterval, 10))
	}
	if info.DataNodeFlushFDParallelismOnDisk > 0 {
		request.addParam("dataNodeFlushFDParallelismOnDisk", strconv.FormatInt(info.DataNodeFlushFDParallelismOnDisk, 10))
	}
	if info.DNNormalExtentDeleteExpire > 0 {
		request.addParam("normalExtentDeleteExpire", strconv.FormatUint(uint64(info.DNNormalExtentDeleteExpire), 10))
	}
	if info.ClientReadVolRate >= 0 {
		request.addParam("clientReadVolRate", strconv.FormatInt(info.ClientReadVolRate, 10))
	}
	if info.ClientWriteVolRate >= 0 {
		request.addParam("clientWriteVolRate", strconv.FormatInt(info.ClientWriteVolRate, 10))
	}
	if info.ClientVolOpRate >= -1 {
		request.addParam("clientVolOpRate", strconv.FormatInt(info.ClientVolOpRate, 10))
	}
	if info.Action != "" {
		request.addParam("action", info.Action)
	}
	if info.ObjectVolActionRate >= -1 {
		request.addParam("objectVolActionRate", strconv.FormatInt(info.ObjectVolActionRate, 10))
	}
	if info.MetaNodeReqRate >= 0 {
		request.addParam("metaNodeReqRate", strconv.FormatInt(info.MetaNodeReqRate, 10))
	}
	if info.MetaNodeReqOpRate != 0 {
		request.addParam("metaNodeReqOpRate", strconv.FormatInt(info.MetaNodeReqOpRate, 10))
	}
	if info.DataNodeRepairTaskCount > 0 {
		request.addParam("dataNodeRepairTaskCount", strconv.FormatInt(info.DataNodeRepairTaskCount, 10))
	}
	if info.DataNodeRepairTaskSSDZone > 0 {
		request.addParam("dataNodeRepairTaskSSDZoneCount", strconv.FormatInt(info.DataNodeRepairTaskSSDZone, 10))
	}
	if info.DataNodeRepairTaskZoneCount >= 0 {
		request.addParam("dataNodeRepairTaskZoneCount", strconv.FormatInt(info.DataNodeRepairTaskZoneCount, 10))
	}
	if info.DataNodeMarkDeleteRate >= 0 {
		request.addParam("markDeleteRate", strconv.FormatInt(info.DataNodeMarkDeleteRate, 10))
	}
	if info.DataNodeReqRate >= 0 {
		request.addParam("dataNodeReqRate", strconv.FormatInt(info.DataNodeReqRate, 10))
	}
	if info.DataNodeReqOpRate >= 0 {
		request.addParam("dataNodeReqOpRate", strconv.FormatInt(info.DataNodeReqOpRate, 10))
	}
	if info.DataNodeReqVolOpRate >= 0 {
		request.addParam("dataNodeReqVolOpRate", strconv.FormatInt(info.DataNodeReqVolOpRate, 10))
	}
	if info.DataNodeReqVolPartRate >= 0 {
		request.addParam(proto.DataNodeReqVolPartRateKey, strconv.FormatInt(info.DataNodeReqVolPartRate, 10))
	}
	if info.DataNodeReqVolOpPartRate >= 0 {
		request.addParam(proto.DataNodeReqVolOpPartRateKey, strconv.FormatInt(info.DataNodeReqVolOpPartRate, 10))
	}
	if info.FlashNodeRate >= 0 {
		request.addParam(proto.FlashNodeRateKey, strconv.FormatInt(info.FlashNodeRate, 10))
	}
	if info.FlashNodeVolRate >= 0 {
		request.addParam(proto.FlashNodeVolRateKey, strconv.FormatInt(info.FlashNodeVolRate, 10))
	}
	if info.ExtentMergeIno != "" {
		request.addParam("extentMergeIno", info.ExtentMergeIno)
	}
	if info.ExtentMergeSleepMs >= 0 {
		request.addParam("extentMergeSleepMs", strconv.FormatInt(info.ExtentMergeSleepMs, 10))
	}
	if info.DnFixTinyDeleteRecordLimit >= 0 {
		request.addParam("fixTinyDeleteRecordKey", strconv.FormatInt(info.DnFixTinyDeleteRecordLimit, 10))
	}
	if info.MetaNodeDumpWaterLevel > 0 {
		request.addParam("metaNodeDumpWaterLevel", strconv.FormatInt(int64(info.MetaNodeDumpWaterLevel), 10))
	}
	if info.MonitorSummarySecond > 0 {
		request.addParam("monitorSummarySec", strconv.FormatUint(info.MonitorSummarySecond, 10))
	}
	if info.MonitorReportSecond > 0 {
		request.addParam("monitorReportSec", strconv.FormatUint(info.MonitorReportSecond, 10))
	}
	if info.RocksDBDiskReservedSpace > 0 {
		request.addParam(proto.RocksDBDiskReservedSpaceKey, strconv.FormatUint(info.RocksDBDiskReservedSpace, 10))
	}
	if info.LogMaxMB > 0 {
		request.addParam(proto.LogMaxMB, strconv.FormatUint(info.LogMaxMB, 10))
	}
	if info.MetaRockDBWalFileMaxMB > 0 {
		request.addParam(proto.MetaRockDBWalFileMaxMB, strconv.FormatUint(info.MetaRockDBWalFileMaxMB, 10))
	}
	if info.MetaRocksWalMemMaxMB > 0 {
		request.addParam(proto.MetaRocksDBWalMemMaxMB, strconv.FormatUint(info.MetaRocksWalMemMaxMB, 10))
	}
	if info.MetaRocksLogMaxMB > 0 {
		request.addParam(proto.MetaRocksDBLogMaxMB, strconv.FormatUint(info.MetaRocksLogMaxMB, 10))
	}
	if info.MetaRocksLogReservedDay > 0 {
		request.addParam(proto.MetaRocksLogReservedDay, strconv.FormatUint(info.MetaRocksLogReservedDay, 10))
	}
	if info.MetaRocksLogReservedCnt > 0 {
		request.addParam(proto.MetaRocksLogReservedCnt, strconv.FormatUint(info.MetaRocksLogReservedCnt, 10))
	}
	if info.MetaRocksFlushWalInterval > 0 {
		request.addParam(proto.MetaRocksWalFlushIntervalKey, strconv.FormatUint(info.MetaRocksFlushWalInterval, 10))
	}
	if info.MetaRocksDisableFlushFlag >= 0 {
		request.addParam(proto.MetaRocksDisableFlushWalKey, strconv.FormatInt(info.MetaRocksDisableFlushFlag, 10))
	}
	if info.MetaRocksWalTTL > 0 {
		request.addParam(proto.MetaRocksWalTTLKey, strconv.FormatUint(info.MetaRocksWalTTL, 10))
	}
	if info.MetaDelEKRecordFileMaxMB > 0 {
		request.addParam(proto.MetaDelEKRecordFileMaxMB, strconv.FormatUint(info.MetaDelEKRecordFileMaxMB, 10))
	}
	if info.MetaTrashCleanInterval > 0 {
		request.addParam(proto.MetaTrashCleanIntervalKey, strconv.FormatUint(info.MetaTrashCleanInterval, 10))
	}
	if info.MetaRaftLogSize > 0 {
		request.addParam(proto.MetaRaftLogSizeKey, strconv.FormatInt(info.MetaRaftLogSize, 10))
	}
	if info.MetaRaftLogCap > 0 {
		request.addParam(proto.MetaRaftLogCapKey, strconv.FormatInt(info.MetaRaftLogCap, 10))
	}
	if info.DataSyncWALEnableState == 0 || info.DataSyncWALEnableState == 1 {
		request.addParam(proto.DataSyncWalEnableStateKey, strconv.FormatInt(info.DataSyncWALEnableState, 10))
	}
	if info.MetaSyncWALEnableState == 0 || info.MetaSyncWALEnableState == 1 {
		request.addParam(proto.MetaSyncWalEnableStateKey, strconv.FormatInt(info.MetaSyncWALEnableState, 10))
	}
	if info.DisableStrictVolZone == 0 || info.DisableStrictVolZone == 1 {
		request.addParam(proto.DisableStrictVolZoneKey, strconv.FormatBool(info.DisableStrictVolZone == 1))
	}
	if info.AutoUpdatePartitionReplicaNum == 0 || info.AutoUpdatePartitionReplicaNum == 1 {
		request.addParam(proto.AutoUpPartitionReplicaNumKey, strconv.FormatBool(info.AutoUpdatePartitionReplicaNum == 1))
	}
	if info.AllocatorMaxUsedFactor > 0 {
		request.addParam(proto.AllocatorMaxUsedFactorKey, strconv.FormatFloat(info.AllocatorMaxUsedFactor, 'f', -1, 64))
	}
	if info.AllocatorMinFreeFactor > 0 {
		request.addParam(proto.AllocatorMinFreeFactorKey, strconv.FormatFloat(info.AllocatorMinFreeFactor, 'f', -1, 64))
	}
	if info.TrashCleanMaxCountEachTime >= 0 {
		request.addParam(proto.TrashItemCleanMaxCountKey, strconv.FormatInt(int64(info.TrashCleanMaxCountEachTime), 10))
	}
	if info.TrashCleanDurationEachTime >= 0 {
		request.addParam(proto.TrashCleanDurationKey, strconv.FormatInt(int64(info.TrashCleanDurationEachTime), 10))
	}
	if info.DeleteMarkDelVolInterval >= 0 {
		request.addParam(proto.DeleteMarkDelVolIntervalKey, strconv.FormatInt(info.DeleteMarkDelVolInterval, 10))
	}
	if info.RemoteCacheBoostEnableState == 0 || info.RemoteCacheBoostEnableState == 1 {
		request.addParam(proto.RemoteCacheBoostEnableKey, strconv.FormatInt(info.RemoteCacheBoostEnableState, 10))
	}
	if info.ClientConnTimeoutUs >= 0 {
		request.addParam(proto.NetConnTimeoutUsKey, strconv.FormatInt(info.ClientConnTimeoutUs, 10))
	}
	if mode := proto.ConsistencyModeFromInt32(info.DataPartitionConsistencyMode); mode.Valid() {
		request.addParam("dataPartitionConsistencyMode", strconv.FormatInt(int64(mode.Int32()), 10))
	}
	if info.DpTimeoutCntThreshold >= 0 {
		request.addParam(proto.DpTimeoutCntThreshold, strconv.FormatInt(int64(info.DpTimeoutCntThreshold), 10))
	}
	if info.ClientReqRecordsReservedCount > 0 {
		request.addParam(proto.ClientReqRecordReservedCntKey, strconv.FormatUint(uint64(info.ClientReqRecordsReservedCount), 10))
	}
	if info.ClientReqRecordsReservedMin > 0 {
		request.addParam(proto.ClientReqRecordReservedMinKey, strconv.FormatUint(uint64(info.ClientReqRecordsReservedMin), 10))
	}
	if info.ClientReqRemoveDupFlag == 0 {
		request.addParam(proto.ClientReqRemoveDupFlagKey, strconv.FormatBool(false))
	} else if info.ClientReqRemoveDupFlag > 0 {
		request.addParam(proto.ClientReqRemoveDupFlagKey, strconv.FormatBool(true))
	}
	request.addParam("volume", info.Volume)
	request.addParam("zoneName", info.ZoneName)
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) ZoneList() (zoneViews []*proto.ZoneView, err error) {
	var request = newAPIRequest(http.MethodGet, proto.GetAllZones)
	var data []byte
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	zoneViews = make([]*proto.ZoneView, 0)
	if err = json.Unmarshal(data, &zoneViews); err != nil {
		return
	}
	return
}

func (api *AdminAPI) GetRegionView(regionName string) (rv *proto.RegionView, err error) {
	var request = newAPIRequest(http.MethodGet, proto.GetRegionView)
	request.addParam("regionName", regionName)
	var data []byte
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	rv = &proto.RegionView{}
	if err = json.Unmarshal(data, rv); err != nil {
		return
	}
	return
}

func (api *AdminAPI) RegionList() (regionViews []*proto.RegionView, err error) {
	var request = newAPIRequest(http.MethodGet, proto.RegionList)
	var data []byte
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	regionViews = make([]*proto.RegionView, 0)
	if err = json.Unmarshal(data, &regionViews); err != nil {
		return
	}
	return
}

func (api *AdminAPI) CreateRegion(regionName string, regionType uint8) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.CreateRegion)
	request.addParam("regionName", regionName)
	request.addParam("regionType", strconv.Itoa(int(regionType)))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) UpdateRegion(regionName string, regionType uint8) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.UpdateRegion)
	request.addParam("regionName", regionName)
	request.addParam("regionType", strconv.Itoa(int(regionType)))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) SetZoneRegion(zoneName, regionName string) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.SetZoneRegion)
	request.addParam("zoneName", zoneName)
	request.addParam("regionName", regionName)
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) SetVolMinRWPartition(volName string, minRwMPNum, minRwDPNum int) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminSetVolMinRWPartition)
	request.addParam("name", volName)
	request.addParam("minWritableMp", strconv.Itoa(minRwMPNum))
	request.addParam("minWritableDp", strconv.Itoa(minRwDPNum))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) GetIdc(name string) (view *proto.IDCView, err error) {
	var request = newAPIRequest(http.MethodGet, proto.GetIDCView)
	request.addParam("name", name)

	data, _, err := api.mc.serveRequest(request)
	if err != nil {
		return
	}
	view = &proto.IDCView{}
	if err = json.Unmarshal(data, &view); err != nil {
		return
	}
	return
}

func (api *AdminAPI) DeleteIdc(name string) (result string, err error) {
	var request = newAPIRequest(http.MethodGet, proto.DeleteDC)
	request.addParam("name", name)

	data, _, err := api.mc.serveRequest(request)
	if err != nil {
		return
	}
	return string(data), nil
}

func (api *AdminAPI) CreateIdc(name string) (result string, err error) {
	var request = newAPIRequest(http.MethodGet, proto.CreateIDC)
	request.addParam("name", name)

	data, _, err := api.mc.serveRequest(request)
	if err != nil {
		return
	}
	return string(data), nil
}

func (api *AdminAPI) IdcList() (views []*proto.IDCView, err error) {
	var request = newAPIRequest(http.MethodGet, proto.IDCList)
	data, _, err := api.mc.serveRequest(request)

	if err != nil {
		return
	}
	views = make([]*proto.IDCView, 0)
	if err = json.Unmarshal(data, &views); err != nil {
		return
	}

	return
}

func (api *AdminAPI) SetIDC(idcName, mediumType, zoneName string) (result string, err error) {
	var request = newAPIRequest(http.MethodGet, proto.SetZoneIDC)
	request.addParam("idcName", idcName)
	request.addParam("mediumType", mediumType)
	request.addParam("zoneName", zoneName)

	data, _, err := api.mc.serveRequest(request)
	if err != nil {
		return
	}
	return string(data), nil
}

func (api *AdminAPI) DataPartitionTransfer(partitionId uint64, address, destAddress string) (result string, err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminTransferDataPartition)
	request.addParam("id", strconv.Itoa(int(partitionId)))
	request.addParam("addr", address)
	request.addParam("destAddr", destAddress)

	data, _, err := api.mc.serveRequest(request)
	if err != nil {
		return
	}
	return string(data), nil
}

func (api *AdminAPI) ListSmartVolumes() (volumes []*proto.SmartVolume, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminSmartVolList)
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}

	volumes = make([]*proto.SmartVolume, 0)
	if err = json.Unmarshal(buf, &volumes); err != nil {
		return
	}
	return
}

func (api *AdminAPI) GetSmartVolume(volName, authKey string) (sv *proto.SmartVolume, err error) {
	var buf []byte
	request := newAPIRequest(http.MethodGet, proto.ClientVol)
	request.addParam("name", volName)
	request.addParam("authKey", authKey)
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	sv = &proto.SmartVolume{}
	if err = json.Unmarshal(buf, &sv); err != nil {
		return
	}
	return
}

func (api *AdminAPI) TransferSmartVolDataPartition(dpId uint64, addr string) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminTransferDataPartition)
	request.addParam("id", strconv.FormatUint(dpId, 10))
	request.addParam("addr", addr)
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}
func (api *AdminAPI) GetEcPartition(volName string, partitionID uint64) (partition *proto.EcPartitionInfo, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminGetEcPartition)
	request.addParam("id", strconv.Itoa(int(partitionID)))
	request.addParam("name", volName)
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	partition = &proto.EcPartitionInfo{}
	err = json.Unmarshal(buf, &partition)
	return
}

func (api *AdminAPI) DecommissionEcPartition(partitionID uint64, nodeAddr string) (data []byte, err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminDecommissionEcPartition)
	request.addParam("id", strconv.FormatUint(partitionID, 10))
	request.addParam("addr", nodeAddr)

	data, _, err = api.mc.serveRequest(request)
	return
}

func (api *AdminAPI) DeleteEcReplica(ecPartitionID uint64, nodeAddr string) (data []byte, err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminDeleteEcReplica)
	request.addParam("id", strconv.FormatUint(ecPartitionID, 10))
	request.addParam("addr", nodeAddr)

	data, _, err = api.mc.serveRequest(request)
	return
}

func (api *AdminAPI) AddEcReplica(ecPartitionID uint64, nodeAddr string) (data []byte, err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminAddEcReplica)
	request.addParam("id", strconv.FormatUint(ecPartitionID, 10))
	request.addParam("addr", nodeAddr)

	data, _, err = api.mc.serveRequest(request)
	return
}

func (api *AdminAPI) SetEcRollBack(ecPartitionID uint64, needDelEc bool) (data []byte, err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminEcPartitionRollBack)
	request.addParam("id", strconv.FormatUint(ecPartitionID, 10))
	request.addParam("delEc", strconv.FormatBool(needDelEc))
	data, _, err = api.mc.serveRequest(request)
	return
}

func (api *AdminAPI) DiagnoseEcPartition() (diagnosis *proto.EcPartitionDiagnosis, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminDiagnoseEcPartition)
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	diagnosis = &proto.EcPartitionDiagnosis{}
	if err = json.Unmarshal(buf, &diagnosis); err != nil {
		return
	}
	return
}

func (api *AdminAPI) UpdateVolumeEcInfo(volName string, ecEnable bool, ecDataNum, ecParityNum int, ecSaveTime, ecWaitTime, ecTimeOut, ecRetryWait int64, ecMaxUnitSize uint64) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminUpdateVolEcInfo)

	request.addParam("name", volName)
	request.addParam("ecMaxUnitSize", strconv.FormatUint(ecMaxUnitSize, 10))
	request.addParam("ecDataNum", strconv.Itoa(ecDataNum))
	request.addParam("ecParityNum", strconv.Itoa(ecParityNum))
	request.addParam("ecSaveTime", strconv.FormatInt(ecSaveTime, 10))
	request.addParam("ecWaitTime", strconv.FormatInt(ecWaitTime, 10))
	request.addParam("ecTimeOut", strconv.FormatInt(ecTimeOut, 10))
	request.addParam("ecRetryWait", strconv.FormatInt(ecRetryWait, 10))
	request.addParam("ecEnable", strconv.FormatBool(ecEnable))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) UpdateEcInfo(ecScrubEnable bool, ecMaxScrubExtents, ecScrubPeriod, maxCodecConcurrent int) (err error) {
	var request = newAPIRequest(http.MethodGet, proto.AdminClusterEcSet)
	request.addParam("ecScrubPeriod", strconv.Itoa(ecScrubPeriod))
	request.addParam("ecMaxScrubExtents", strconv.Itoa(ecMaxScrubExtents))
	request.addParam("maxCodecConcurrent", strconv.Itoa(maxCodecConcurrent))
	request.addParam("ecScrubEnable", strconv.FormatBool(ecScrubEnable))
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return
}

func (api *AdminAPI) GetCanMigrateDataPartitions() (partitions []*proto.DataPartitionResponse, err error) {
	var data []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminCanMigrateDataPartitions)
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	partitions = make([]*proto.DataPartitionResponse, 0)
	if err = json.Unmarshal(data, &partitions); err != nil {
		return
	}
	return
}

func (api *AdminAPI) GetCanDelDataPartitions() (partitions []*proto.DataPartitionResponse, err error) {
	var data []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminCanDelDataPartitions)
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	partitions = make([]*proto.DataPartitionResponse, 0)
	if err = json.Unmarshal(data, &partitions); err != nil {
		return
	}
	return
}

func (api *AdminAPI) DeleteDpAlreadyEc(ecPartitionID uint64) string {
	var request = newAPIRequest(http.MethodGet, proto.AdminDelDpAlreadyEc)
	request.addParam("id", strconv.FormatUint(ecPartitionID, 10))
	data, _, err := api.mc.serveRequest(request)
	if err != nil {
		return fmt.Sprintf("DeleteDpAlreadyEc fail:%v\n", err)
	}
	return string(data)
}

func (api *AdminAPI) MigrateEcById(ecPartitionID uint64, test bool) string {
	var request = newAPIRequest(http.MethodGet, proto.AdminDpMigrateEc)
	request.addParam("id", strconv.FormatUint(ecPartitionID, 10))
	request.addParam("test", strconv.FormatBool(test))
	data, _, err := api.mc.serveRequest(request)
	if err != nil {
		return fmt.Sprintf("MigrateEcById fail: %v\n", err)
	}
	return string(data)
}

func (api *AdminAPI) StopMigratingByDataPartition(dataPartitionID uint64) string {
	var request = newAPIRequest(http.MethodGet, proto.AdminDpStopMigrating)
	request.addParam("id", strconv.FormatUint(dataPartitionID, 10))
	data, _, err := api.mc.serveRequest(request)
	if err != nil {
		return fmt.Sprintf("StopMigratingByDataPartition fail:%v\n", err)
	}
	return string(data)
}

func (api *AdminAPI) ListCompactVolumes() (volumes []*proto.CompactVolume, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminCompactVolList)
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}

	volumes = make([]*proto.CompactVolume, 0)
	if err = json.Unmarshal(buf, &volumes); err != nil {
		return
	}
	return
}

func (api *AdminAPI) SetCompact(volName, compactTag, authKey string) (result string, err error) {
	var data []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminCompactVolSet)
	request.addParam("name", volName)
	request.addParam("compactTag", compactTag)
	request.addParam("authKey", authKey)
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return string(data), nil
}

func (api *AdminAPI) SetVolChildFileMaxCount(volName string, maxCount uint32) (result string, err error) {
	var data []byte
	var req = newAPIRequest(http.MethodGet, proto.AdminSetVolChildMaxCnt)
	req.addParam(proto.NameKey, volName)
	req.addParam(proto.ChildFileMaxCountKey, strconv.FormatUint(uint64(maxCount), 10))
	if data, _, err = api.mc.serveRequest(req); err != nil {
		return
	}
	return string(data), nil
}

func (api *AdminAPI) CreateFlashGroup(slots string) (fgView proto.FlashGroupAdminView, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminCreateFlashGroup)
	request.addParam("slots", slots)
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	if err = json.Unmarshal(buf, &fgView); err != nil {
		return
	}
	return
}

func (api *AdminAPI) SetFlashGroup(flashGroupID uint64, isActive bool) (fgView proto.FlashGroupAdminView, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminSetFlashGroup)
	request.addParam("id", strconv.FormatUint(flashGroupID, 10))
	request.addParam("enable", strconv.FormatBool(isActive))
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	if err = json.Unmarshal(buf, &fgView); err != nil {
		return
	}
	return
}

func (api *AdminAPI) GetFlashGroup(flashGroupID uint64) (fgView proto.FlashGroupAdminView, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminGetFlashGroup)
	request.addParam("id", strconv.FormatUint(flashGroupID, 10))
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	if err = json.Unmarshal(buf, &fgView); err != nil {
		return
	}
	return
}

func (api *AdminAPI) RemoveFlashGroup(flashGroupID uint64) (result string, err error) {
	var data []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminRemoveFlashGroup)
	request.addParam("id", strconv.FormatUint(flashGroupID, 10))
	if data, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return string(data), nil
}

func (api *AdminAPI) FlashGroupAddFlashNode(flashGroupID uint64, count int, zoneName, addr string) (fgView proto.FlashGroupAdminView, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminFlashGroupAddFlashNode)
	request.addParam("id", strconv.FormatUint(flashGroupID, 10))
	request.addParam("count", strconv.Itoa(count))
	request.addParam("zoneName", zoneName)
	request.addParam("addr", addr)
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	if err = json.Unmarshal(buf, &fgView); err != nil {
		return
	}
	return
}

func (api *AdminAPI) FlashGroupRemoveFlashNode(flashGroupID uint64, count int, zoneName, addr string) (fgView proto.FlashGroupAdminView, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminFlashGroupRemoveFlashNode)
	request.addParam("id", strconv.FormatUint(flashGroupID, 10))
	request.addParam("count", strconv.Itoa(count))
	request.addParam("zoneName", zoneName)
	request.addParam("addr", addr)
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	if err = json.Unmarshal(buf, &fgView); err != nil {
		return
	}
	return
}

func (api *AdminAPI) ListFlashGroups(isActive, listAllStatus bool) (fgView proto.FlashGroupsAdminView, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.AdminListFlashGroups)
	if !listAllStatus {
		request.addParam("enable", strconv.FormatBool(isActive))
	}
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	if err = json.Unmarshal(buf, &fgView); err != nil {
		return
	}
	return
}

func (api *AdminAPI) ClientFlashGroups() (fgView *proto.FlashGroupView, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.ClientFlashGroups)
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	fgView = &proto.FlashGroupView{}
	if err = json.Unmarshal(buf, fgView); err != nil {
		return
	}
	return
}

func (api *AdminAPI) GetAllFlashNodes(getAllFlashNodes bool) (zoneFlashNodes map[string][]*proto.FlashNodeViewInfo, err error) {
	var buf []byte
	var request = newAPIRequest(http.MethodGet, proto.GetAllFlashNodes)
	request.addParam("getAllFlashNodes", strconv.FormatBool(getAllFlashNodes))
	if buf, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	if err = json.Unmarshal(buf, &zoneFlashNodes); err != nil {
		return
	}
	return
}

// state: "true" or "false"
func (api *AdminAPI) SetNodeToOfflineState(addrList []string, nodeType, zoneName, state string) (err error) {
	var body []byte
	if body, err = json.Marshal(addrList); err != nil {
		return
	}
	var request = newAPIRequest(http.MethodPost, proto.AdminSetNodeStateByAddr)
	request.addParam("nodeType", nodeType)
	request.addParam("zoneName", zoneName)
	request.addParam("state", state)
	request.addBody(body)
	if _, _, err = api.mc.serveRequest(request); err != nil {
		return
	}
	return nil
}

func (api *AdminAPI) SetNodeSetCapacity(capacity int) (result string, err error) {
	var data []byte
	var req = newAPIRequest(http.MethodPost, proto.AdminSetNodeSetCapacity)
	req.addParam(proto.NodeSetCapacityKey, strconv.Itoa(capacity))
	if data, _, err = api.mc.serveRequest(req); err != nil {
		return
	}
	return string(data), nil
}
