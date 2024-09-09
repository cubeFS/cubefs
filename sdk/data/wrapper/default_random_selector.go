// Copyright 2020 The CubeFS Authors.
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

package wrapper

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/cubefs/cubefs/util/log"
)

const (
	DefaultRandomSelectorName = "default"
)

func init() {
	_ = RegisterDataPartitionSelector(DefaultRandomSelectorName, newDefaultRandomSelector)
}

func newDefaultRandomSelector(_ string) (selector DataPartitionSelector, e error) {
	selector = &DefaultRandomSelector{
		localLeaderPartitions: make([]*DataPartition, 0),
		partitions:            make([]*DataPartition, 0),
	}
	return
}

type DefaultRandomSelector struct {
	sync.RWMutex
	localLeaderPartitions []*DataPartition
	partitions            []*DataPartition
	removeDpMutex         sync.Mutex
}

func (s *DefaultRandomSelector) Name() string {
	return DefaultRandomSelectorName
}

func (s *DefaultRandomSelector) Refresh(partitions []*DataPartition) (err error) {
	var localLeaderPartitions []*DataPartition
	for i := 0; i < len(partitions); i++ {
		if strings.Split(partitions[i].Hosts[0], ":")[0] == LocalIP {
			localLeaderPartitions = append(localLeaderPartitions, partitions[i])
		}
	}

	s.Lock()
	defer s.Unlock()

	s.localLeaderPartitions = localLeaderPartitions
	s.partitions = partitions
	return
}

func (s *DefaultRandomSelector) Select(exclude map[string]struct{}) (dp *DataPartition, err error) {
	dp = s.getLocalLeaderDataPartition(exclude)
	if dp != nil {
		log.LogDebugf("Select: select dp[%v] address[%p] from LocalLeaderDataPartition", dp, dp)
		return dp, nil
	}

	s.RLock()
	partitions := s.partitions
	log.LogDebugf("Select: len(s.partitions)=%v\n", len(s.partitions))
	s.RUnlock()

	dp = s.getRandomDataPartition(partitions, exclude)

	if dp != nil {
		return dp, nil
	}
	log.LogErrorf("DefaultRandomSelector: no writable data partition with %v partitions and exclude(%v)",
		len(partitions), exclude)
	return nil, fmt.Errorf("no writable data partition")
}

func (s *DefaultRandomSelector) RemoveDP(partitionID uint64) {
	s.removeDpMutex.Lock()
	defer s.removeDpMutex.Unlock()

	s.RLock()
	rwPartitionGroups := s.partitions
	localLeaderPartitions := s.localLeaderPartitions
	log.LogDebugf("RemoveDP: partitionID[%v], len(s.partitions)=%v len(s.localLeaderPartitions)=%v\n", partitionID, len(s.partitions), len(s.localLeaderPartitions))
	s.RUnlock()

	var i int
	for i = 0; i < len(rwPartitionGroups); i++ {
		if rwPartitionGroups[i].PartitionID == partitionID {
			log.LogDebugf("RemoveDP: found partitionID[%v] in rwPartitionGroups. dp[%v] address[%p]\n", partitionID, rwPartitionGroups[i], rwPartitionGroups[i])
			break
		}
	}
	if i >= len(rwPartitionGroups) {
		log.LogDebugf("RemoveDP: not found partitionID[%v] in rwPartitionGroups", partitionID)
		return
	}

	newRwPartition := make([]*DataPartition, 0)
	newRwPartition = append(newRwPartition, rwPartitionGroups[:i]...)
	newRwPartition = append(newRwPartition, rwPartitionGroups[i+1:]...)

	defer func() {
		s.Lock()
		s.partitions = newRwPartition
		log.LogDebugf("RemoveDP: finish, partitionID[%v], len(s.partitions)=%v\n", partitionID, len(s.partitions))
		s.Unlock()
	}()

	for i = 0; i < len(localLeaderPartitions); i++ {
		if localLeaderPartitions[i].PartitionID == partitionID {
			log.LogDebugf("RemoveDP: found partitionID[%v] in localLeaderPartitions. dp[%v] address[%p]\n", partitionID, localLeaderPartitions[i], localLeaderPartitions[i])
			break
		}
	}
	if i >= len(localLeaderPartitions) {
		log.LogDebugf("RemoveDP: not found partitionID[%v] in localLeaderPartitions", partitionID)
		return
	}
	newLocalLeaderPartitions := make([]*DataPartition, 0)
	newLocalLeaderPartitions = append(newLocalLeaderPartitions, localLeaderPartitions[:i]...)
	newLocalLeaderPartitions = append(newLocalLeaderPartitions, localLeaderPartitions[i+1:]...)

	s.Lock()
	s.localLeaderPartitions = newLocalLeaderPartitions
	s.Unlock()
	log.LogDebugf("RemoveDP: finish, partitionID[%v], len(s.localLeaderPartitions)=%v\n", partitionID, len(s.localLeaderPartitions))
}

func (s *DefaultRandomSelector) Count() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.partitions)
}

func (s *DefaultRandomSelector) getLocalLeaderDataPartition(exclude map[string]struct{}) *DataPartition {
	s.RLock()
	localLeaderPartitions := s.localLeaderPartitions
	log.LogDebugf("getLocalLeaderDataPartition: len(s.localLeaderPartitions)=%v\n", len(s.localLeaderPartitions))
	s.RUnlock()
	return s.getRandomDataPartition(localLeaderPartitions, exclude)
}

func (s *DefaultRandomSelector) getRandomDataPartition(partitions []*DataPartition, exclude map[string]struct{}) (
	dp *DataPartition,
) {
	length := len(partitions)
	if length == 0 {
		return nil
	}

	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(length)
	dp = partitions[index]
	if !isExcluded(dp, exclude) {
		log.LogDebugf("DefaultRandomSelector: select dp[%v] address[%p], index %v", dp, dp, index)
		return dp
	}

	log.LogWarnf("DefaultRandomSelector: first random partition was excluded, get partition from others")

	var currIndex int
	for i := 0; i < length; i++ {
		currIndex = (index + i) % length
		if !isExcluded(partitions[currIndex], exclude) {
			log.LogDebugf("DefaultRandomSelector: select dp[%v], index %v", partitions[currIndex], currIndex)
			return partitions[currIndex]
		}
	}
	return nil
}

func (s *DefaultRandomSelector) GetAllDp() (dps []*DataPartition) {
	s.RLock()
	defer s.RUnlock()
	dps = make([]*DataPartition, len(s.partitions))
	copy(dps, s.partitions)
	return
}