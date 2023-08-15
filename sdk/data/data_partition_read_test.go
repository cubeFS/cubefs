package data

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/cubefs/cubefs/sdk/common"
	"github.com/cubefs/cubefs/sdk/master"
	"github.com/cubefs/cubefs/sdk/meta"
	"github.com/cubefs/cubefs/util/exporter"
)

var (
	readFilePath = "/cfs/mnt/testread.txt"
)

func TestLeaderRead(t *testing.T) {
	readFile, err := os.Create(readFilePath)
	if err != nil && !os.IsExist(err) {
		t.Fatalf("create file failed: err(%v) file(%v)", err, readFilePath)
	}
	defer readFile.Close()
	writeData := "testLeaderRead"
	writeBytes := []byte(writeData)
	writeOffset := int64(0)
	for i := 0; i < 1024; i++ {
		var writeLen int
		if writeLen, err = readFile.WriteAt(writeBytes, writeOffset); err != nil {
			t.Fatalf("write file failed: err(%v) expect write length(%v) but (%v)", err, len(writeBytes), writeLen)
		}
		writeOffset += int64(writeLen)
	}
	if err = readFile.Sync(); err != nil {
		t.Errorf("sync file failed: err(%v) file(%v)", err, readFilePath)
	}
	readBytes := make([]byte, len(writeBytes))
	readOffset := int64(0)
	for i := 0; i < 1024; i++ {
		var readLen int
		if readLen, err = readFile.ReadAt(readBytes, readOffset); err != nil {
			t.Errorf("read file failed: err(%v) expect read length(%v) but (%v)", err, len(readBytes), readLen)
		}
		if string(readBytes) != writeData {
			t.Errorf("read file failed: err(%v) expect read data(%v) but (%v)", err, writeData, string(readBytes))
		}
		readOffset += int64(readLen)
	}
}

func TestNearRead(t *testing.T) {
	masters := strings.Split(ltptestMaster, ",")
	testMc := master.NewMasterClient(masters, false)
	volumeSimpleInfo, _ := testMc.AdminAPI().GetVolumeSimpleInfo("ltptest")
	if err := testMc.AdminAPI().UpdateVolume("ltptest", 30, 3, 3, 30, 1,
		true, false, true, false, false, false, false, false, false, calcAuthKey("ltptest"),
		"default", "0,0", "", 0, 0, 60, volumeSimpleInfo.CompactTag, 0, 0, 0, 0, 0, exporter.UMPCollectMethodUnknown, -1, -1, false,
		"", false, false, 0, false); err != nil {
		t.Fatalf("update followerRead and nearRead to 'true' failed: err(%v) vol(ltptest)", err)
	}
	dataWrapper, err := NewDataPartitionWrapper(ltptestVolume, masters)
	if err != nil {
		t.Fatalf("NewDataPartitionWrapper: err(%v) vol(%v) master addr(%v)", err, ltptestVolume, ltptestMaster)
	}
	dataWrapper.SetNearRead(true)
	dataWrapper.updateDataPartition(false)

	dataWrapper.Lock()
	dataWrapper.partitions.Range(func(key, value interface{}) bool {
		dp := value.(*DataPartition)
		nearHosts := dataWrapper.sortHostsByDistance(dp.Hosts)
		sc := NewStreamConn(dp, true)
		fmt.Println("StreamConn String: ", sc)
		if sc.currAddr != nearHosts[0] {
			t.Errorf("NearRead: expect current nearest address(%v) but(%v)", nearHosts, sc.currAddr)
		}
		return true
	})
	dataWrapper.Unlock()
	// verify sort near hosts
	originLocalIP := LocalIP
	LocalIP = "192.168.0.2"
	testsForNearRead := []struct {
		name    string
		hosts   []string
		nearest string
	}{
		{
			name: "test01",
			hosts: []string{
				LocalIP,
				"192.168.0.11",
				"192.168.0.12",
			},
			nearest: LocalIP,
		},
		{
			name: "test02",
			hosts: []string{
				LocalIP,
			},
			nearest: LocalIP,
		},
		{
			name: "test03",
			hosts: []string{
				LocalIP,
				"192.168.0.11",
			},
			nearest: LocalIP,
		},
		{
			name: "test04",
			hosts: []string{
				"192.168.1.11",
				"192.168.0.11",
				"192.168.0.21",
			},
			nearest: "192.168.0.11",
		},
		{
			name: "test05",
			hosts: []string{
				"192.169.2.2",
				"193.168.0.2",
				"192.168.1.1",
			},
			nearest: "192.168.1.1",
		},
		{
			name: "test06",
			hosts: []string{
				"192.169.0.2",
				"192.168.3.3",
				"193.167.0.2",
			},
			nearest: "192.168.3.3",
		},
	}
	for _, tt := range testsForNearRead {
		t.Run(tt.name, func(t *testing.T) {
			beforeHosts := make([]string, len(tt.hosts))
			copy(beforeHosts, tt.hosts)
			nearHosts := dataWrapper.sortHostsByDistance(tt.hosts)
			afterHosts := tt.hosts
			fmt.Printf("NearRead: before hosts(%v) after(%v) near(%v)\n", beforeHosts, afterHosts, nearHosts)
			if len(beforeHosts) != len(afterHosts) {
				t.Errorf("NearRead: hosts changed, expect hosts(%v) but(%v)", beforeHosts, afterHosts)
			}
			for i := 0; i < len(beforeHosts); i++ {
				if beforeHosts[i] != afterHosts[i] {
					t.Errorf("NearRead: hosts changed, expect hosts(%v) but(%v)", beforeHosts, afterHosts)
					break
				}
			}
			if nearHosts[0] != tt.nearest {
				t.Errorf("NearRead: sort hosts error, expect nearest(%v) but(%v)", tt.nearest, nearHosts[0])
			}
		})
	}
	dataWrapper.Stop()
	LocalIP = originLocalIP
	if err = testMc.AdminAPI().UpdateVolume("ltptest", 30, 3, 3, 30, 1,
		false, false, false, false, false, false, false, false, false, calcAuthKey("ltptest"),
		"default", "0,0", "", 0, 0, 60, volumeSimpleInfo.CompactTag, 0, 0, 0, 0, 0, exporter.UMPCollectMethodUnknown, -1, -1, false,
		"", false, false, 0, false); err != nil {
		t.Errorf("update followerRead and nearRead to 'false' failed: err(%v) vol(ltptest)", err)
	}
}

func TestConsistenceRead(t *testing.T) {
	var (
		mw  *meta.MetaWrapper
		ec  *ExtentClient
		err error
	)
	if mw, err = meta.NewMetaWrapper(&meta.MetaConfig{
		Volume:        ltptestVolume,
		Masters:       strings.Split(ltptestMaster, ","),
		ValidateOwner: true,
		Owner:         ltptestVolume,
	}); err != nil {
		t.Fatalf("NewMetaWrapper failed: err(%v) vol(%v)", err, ltptestVolume)
	}
	if ec, err = NewExtentClient(&ExtentConfig{
		Volume:            ltptestVolume,
		Masters:           strings.Split(ltptestMaster, ","),
		FollowerRead:      false,
		OnInsertExtentKey: mw.InsertExtentKey,
		OnGetExtents:      mw.GetExtents,
		OnTruncate:        mw.Truncate,
		TinySize:          NoUseTinyExtent,
	}, nil); err != nil {
		t.Fatalf("NewExtentClient failed: err(%v) vol(%v)", err, ltptestVolume)
	}

	var fInfo os.FileInfo
	if fInfo, err = os.Stat(readFilePath); err != nil {
		t.Fatalf("stat file: err(%v) file(%v)", err, readFilePath)
	}
	sysStat := fInfo.Sys().(*syscall.Stat_t)
	streamMap := ec.streamerConcurrentMap.GetMapSegment(sysStat.Ino)
	streamer := NewStreamer(ec, sysStat.Ino, streamMap, false, false)
	if err = streamer.GetExtents(context.Background()); err != nil {
		t.Fatalf("refresh extent: err(%v) inode(%v)", err, sysStat.Ino)
	}

	eks := streamer.extents.List()
	for _, ek := range eks {
		data := make([]byte, ek.Size)
		req := NewExtentRequest(ek.FileOffset, int(ek.Size), data, 0, uint64(ek.Size), &ek)
		reqPacket := common.NewReadPacket(context.Background(), &ek, int(ek.ExtentOffset), req.Size, streamer.inode, req.FileOffset, false)
		partition, getErr := streamer.client.dataWrapper.GetDataPartition(ek.PartitionId)
		if getErr != nil {
			t.Errorf("GetDataPartition: err(%v) pid(%v)", getErr, ek.PartitionId)
			continue
		}
		// leader read
		sc, leaderReadBytes, leadErr := partition.LeaderRead(reqPacket, req)
		if leadErr != nil {
			t.Errorf("LeaderRead: err(%v) sc(%v) reqPacket(%v) req(%v)", leadErr, sc, reqPacket, req)
			continue
		}
		leadData := req.Data
		req.Data = make([]byte, ek.Size)
		// consistence read
		consisReadBytes, consisErr := partition.ReadConsistentFromHosts(sc, reqPacket, req)
		if consisErr != nil {
			t.Errorf("ReadConsistent: err(%v) sc(%v) reqPacket(%v) req(%v)", consisErr, sc, reqPacket, req)
			continue
		}
		// compare
		if leaderReadBytes != consisReadBytes || string(leadData) != string(req.Data) {
			t.Fatalf("data is not consistent: leaderBytes(%v) consistenBytes(%v) leaderData(%v) consistenData(%v)",
				leaderReadBytes, consisReadBytes, string(leadData), string(req.Data))
		}
		fmt.Printf("read: leaderBytes(%v) consistenBytes(%v) leaderData(%v) consistenData(%v)\n",
			leaderReadBytes, consisReadBytes, len(leadData), len(req.Data))
	}
	// close
	streamer.done <- struct{}{}
	if err = ec.Close(context.Background()); err != nil {
		t.Errorf("Close ExtentClient failed: err(%v) vol(%v)", err, ltptestVolume)
	}
	if err = mw.Close(); err != nil {
		t.Errorf("Close MetaWrapper failed: err(%v) vol(%v)", err, ltptestVolume)
	}
}

func calcAuthKey(key string) (authKey string) {
	h := md5.New()
	_, _ = h.Write([]byte(key))
	cipherStr := h.Sum(nil)
	return strings.ToLower(hex.EncodeToString(cipherStr))
}
