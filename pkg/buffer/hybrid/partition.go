// Copyright 2022 The casbin-neo Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hybrid

import (
	"github.com/casbin-mesh/neo/pkg/primitive/list"
	"sync"
)

type (
	PartitionNum uint8
	PID          uint64
	Partition    struct {
		ioHash  sync.Map
		ioMutex sync.Mutex
		// ---------------------------------
		coolingMutex sync.Mutex
		coolingQueue list.List[*BufferFrame]
		// ---------------------------------
		coolingBfsCounter uint64
		freeBFsLimit      uint64
		coolingBFsLimit   uint64
		dramFreeList      FreeList
		// ---------------------------------
		pidDistance uint64
		pIDsMutex   sync.Mutex
		// freedPIDs
		// some tricks for vector-like usages https://github.com/golang/go/wiki/SliceTricks
		freedPIDs []PID
		nextPid   PID
	}
)

const (
	Partition1   = 1 << 0
	Partition2   = 1 << 1
	Partition4   = 1 << 2
	Partition8   = 1 << 3
	Partition16  = 1 << 4
	Partition32  = 1 << 5
	Partition64  = 1 << 6
	Partition128 = 1 << 7
)

func newPartition(firstPid, pidDistance, freeBFsLimit, coolingBFsLimit uint64) *Partition {
	return &Partition{
		//TODO: init io_ht
		coolingBFsLimit: coolingBFsLimit,
		freeBFsLimit:    freeBFsLimit,
		pidDistance:     pidDistance,
		nextPid:         PID(firstPid),
	}
}

// freePage push freed PID into freedPIDs
func (p *Partition) freePage(pid PID) {
	p.pIDsMutex.Lock()
	defer p.pIDsMutex.Unlock()
	p.freedPIDs = append(p.freedPIDs, pid)
}

// allocatedPages allocated Page count
func (p *Partition) allocatedPages() uint64 {
	return uint64(p.nextPid) / p.pidDistance
}

// nextPID generate next PID
func (p *Partition) nextPID() (pid PID) {
	p.pIDsMutex.Lock()
	defer p.pIDsMutex.Unlock()

	// consume freedPIDs first
	if len(p.freedPIDs) > 0 {
		// pop back
		pid, p.freedPIDs = p.freedPIDs[len(p.freedPIDs)-1], p.freedPIDs[:len(p.freedPIDs)-1]
	} else {
		pid = p.nextPid
		p.nextPid += PID(p.pidDistance)
		// TODO: check PID <= upper bound
	}
	return
}
