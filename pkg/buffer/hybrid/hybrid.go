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
	"github.com/casbin-mesh/neo/pkg/x"
	"golang.org/x/sys/unix"
	"math"
	"math/bits"
	"runtime"
	"unsafe"
)

type BufferManager struct {
	dramPoolSize  uint64 // total number of buffer frames
	fd            int    // file descriptor
	partitions    []Partition
	partitionMask uint64

	mmapMemRef []byte // for release ref (i.e munmap)
	bfs        []BufferFrame
}

type Options struct {
	DramSize               float64 // size dram, GiB
	PartitionNum           PartitionNum
	freeFramePercentage    uint64 // range 1-100
	coolingFramePercentage uint64 // range 1-100
	fd                     int    // file descriptor
}

const (
	SafetyPages = 10 // prevent segfaults
)

func New(opts *Options) *BufferManager {

	if opts == nil {
		// set default
	}

	dramPoolPageSize := uint64(math.Ceil(opts.DramSize * 1024 * 1024 * 1024 / float64(BufferFrameSize)))

	dramTotalSize := (dramPoolPageSize + SafetyPages) * BufferFrameSize

	// by default, only 1 partition
	partitionCount := Partition1
	if opts.PartitionNum > 0 {
		partitionCount = 1 << bits.TrailingZeros8(uint8(opts.PartitionNum))
	}

	vm, _ := unix.Mmap(-1, 0, int(dramTotalSize), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_ANON|unix.MAP_PRIVATE)

	// TODO: do we need to keep it?
	runtime.KeepAlive(vm)

	// reinterpret_cast
	//var bfs []BufferFrame
	//sh := (*reflect.SliceHeader)(unsafe.Pointer(&bfs))
	//sh.Data = uintptr(unsafe.Pointer(&vm[0]))
	//sh.Cap=
	bfHead := (*BufferFrame)(unsafe.Pointer(&vm[0]))
	bfs := unsafe.Slice(bfHead, dramPoolPageSize)

	partitions := make([]Partition, partitionCount)

	freeBFsLimit := math.Ceil(float64(opts.freeFramePercentage) * float64(dramPoolPageSize) / 100 / float64(partitionCount))
	coolingBFsUpperBound := math.Ceil(float64(opts.coolingFramePercentage) * float64(dramPoolPageSize) / 100 / float64(partitionCount))

	// init partitions
	for i := 0; i < partitionCount; i++ {
		partitions[i] = *newPartition(uint64(i), uint64(partitionCount), uint64(freeBFsLimit), uint64(coolingBFsUpperBound))
	}

	// init virtual memory
	x.ParallelRange(dramTotalSize, func(begin, end uint64) {
		//call memset
		x.Memset(vm[begin:end], 0)
	})
	// init partitions
	x.ParallelRange(dramPoolPageSize, func(begin, end uint64) {
		pIdx := 0
		for i := begin; i < end; i++ {
			// vm alloc pages
			bf := &bfs[i]
			partitions[pIdx].dramFreeList.Push(bf)
			pIdx = (pIdx + 1) % partitionCount
		}
	})

	bm := &BufferManager{
		dramPoolSize:  dramPoolPageSize,
		mmapMemRef:    vm,
		partitions:    partitions,
		partitionMask: uint64(len(partitions) - 1),
		bfs:           bfs,
		fd:            opts.fd,
	}

	return bm
}

func (bm *BufferManager) Close() error {

	return nil
}

func (bm *BufferManager) randomPartition() *Partition {
	randPartitionIdx := x.GetRand(0, len(bm.partitions))
	return &bm.partitions[randPartitionIdx]
}

func (bm *BufferManager) getPartitionID(pid PID) uint64 {
	return uint64(pid) & bm.partitionMask
}

func (bm *BufferManager) getPartition(pid PID) *Partition {
	return &bm.partitions[bm.getPartitionID(pid)]
}

// AllocatePage return a write-locked BufferFrame
func (bm *BufferManager) AllocatePage() (bf *BufferFrame) {
	partition := bm.randomPartition()
	pid := partition.nextPID()
	bf = partition.dramFreeList.Pop()
	// Initialize Buffer Frame
	bf.latch.Lock()
	bf.pid = pid
	bf.state = HOT
	bf.lastWrittenGSN = 0

	// TODO: check pid == dram pool size
	return
}

// ReclaimPage reclaim BufferFrame content.
//
// Pre: bf is exclusively locked
//
// ATTENTION: this function unlocks it !!
func (bm *BufferManager) ReclaimPage(bf *BufferFrame) {
	partition := bm.getPartition(bf.pid)
	partition.freePage(bf.pid)
	bf.reset()
	bf.latch.Unlock()
	partition.dramFreeList.Push(bf)
}

func (bm *BufferManager) readPageSync(pid uint64, buf []byte) {
	bytesLeft := PageSize
	for bytesLeft > 0 {
		bytesRead, _ :=
			unix.Pread(
				bm.fd,
				buf,
				int64(int(pid)*PageSize+(PageSize-bytesLeft)),
			)
		bytesLeft -= bytesRead
	}
}

func (bm *BufferManager) writeAllBufferFrames() {
	// TODO: stop background gc
	x.ParallelRange(bm.dramPoolSize, func(begin, end uint64) {
		page := Page{}
		for i := begin; i < end; i++ {
			bf := &bm.bfs[i]
			bf.latch.Lock()
			if !bf.isFree() {
				page.dType = bf.dType
				page.magicDebug = uint64(bf.pid)
				//TODO: checkpoint
				_, err := unix.Pwrite(
					bm.fd,
					unsafe.Slice((*byte)(unsafe.Pointer(&page)), PageSize),
					int64(bf.pid*PageSize))
				//TODO: handle errors
				if err != nil {
					panic(err)
				}
			}
			bf.latch.Unlock()
		}
	})
}
