package master

import (
	"fmt"
	"github.com/cubefs/cubefs/proto"
	"testing"
)

func Test_EcPartition(t *testing.T) {
	/*server.cluster.checkDataNodeHeartbeat()
	server.cluster.checkMetaNodeHeartbeat()
	server.cluster.checkEcNodeHeartbeat()
	partitionId := uint64(1)
	dp, err := server.cluster.getDataPartitionByID(partitionId)
	if err == nil {
		err = server.cluster.ecMigrateById(partitionId, true)
		if !assert.NoErrorf(t, err, "ecMigrateById err(%v)", err) {
			return
		}
	}

	time.Sleep(5 * time.Second)
	server.cluster.checkEcDataPartitions()

	if !assert.Greater(t, len(commonVol.ecDataPartitions.partitions), 0, "getEcDataPartition no ecdp") {
		return
	}
	for _, partition := range commonVol.ecDataPartitions.partitions {
		getEcPartition(partition.PartitionID, t)
		dp.EcMigrateStatus = proto.FinishEC
		decommissionEcPartition(partition, t)
		break
	}*/
}

func getEcPartition(id uint64, t *testing.T) {
	reqURL := fmt.Sprintf("%v%v?id=%v",
		hostAddr, proto.AdminGetEcPartition, id)
	process(reqURL, t)
}

func decommissionEcPartition(ecdp *EcDataPartition, t *testing.T) {
	offlineAddr := ecdp.Hosts[0]
	reqURL := fmt.Sprintf("%v%v?name=%v&id=%v&addr=%v",
		hostAddr, proto.AdminDecommissionEcPartition, ecdp.VolName, ecdp.PartitionID, offlineAddr)
	process(reqURL, t)
}
