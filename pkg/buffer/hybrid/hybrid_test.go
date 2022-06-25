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
	"fmt"
	"github.com/stretchr/testify/assert"
	"sort"
	"sync"
	"testing"
	"unsafe"
)

func Test_Hybrid(t *testing.T) {
	v1 := float64(1)
	v2 := int64(2)
	fmt.Printf("%b\n", unsafe.Pointer(&v1))
	fmt.Printf("%b\n", unsafe.Pointer(&v2))
	fmt.Printf("%b\n", uintptr(unsafe.Pointer(&v1)))
	fmt.Printf("%b\n", uintptr(unsafe.Pointer(&v2)))
	fmt.Printf("%v\n", uintptr(unsafe.Pointer(&v2))-uintptr(unsafe.Pointer(&v1)))
}

func TestBufferManager_New(t *testing.T) {
	bfm := New(&Options{
		DramSize:               0.001, // 1MB
		PartitionNum:           1,
		freeFramePercentage:    10, // 10%
		coolingFramePercentage: 1,  // 1%
	})
	assert.NotNil(t, bfm)
}

func newBufferManager(fd int, num PartitionNum) *BufferManager {
	bfm := New(&Options{
		DramSize:               0.001, // 1MB
		PartitionNum:           num,
		freeFramePercentage:    10, // 10%
		coolingFramePercentage: 1,  // 1%
		fd:                     fd,
	})
	return bfm
}

type BufferFrames []*BufferFrame

func (b BufferFrames) Len() int {
	return len(b)
}

func (b BufferFrames) Less(i, j int) bool {
	return b[i].pid < b[j].pid
}

func (b BufferFrames) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func AssertAllocatedPages(t *testing.T, allocated BufferFrames, partitionCount, pidStart int, pageCount int) {
	assert.Equal(t, pageCount-pidStart, len(allocated))
	allocatedMap := map[int]BufferFrames{}

	for i := 0; i < len(allocated); i++ {
		bf := allocated[i]
		idx := int(bf.pid) % partitionCount
		allocatedMap[idx] = append(allocatedMap[idx], allocated[i])
	}

	// assert each partitions
	for _, frames := range allocatedMap {
		// next.pid = previous.pid + partitionCount
		for i, frame := range frames {
			exp := PID(int(frames[0].pid) + i*partitionCount)
			assert.Equal(t, exp, frame.pid, "iter:%d, should be :%d,but got %d\n", i, exp, frame.pid)
		}
	}
}

func TestBufferManager_AllocatePage(t *testing.T) {
	t.Run("single thread test", func(t *testing.T) {
		bm := newBufferManager(-1, Partition1)
		// allocate a new page
		bf := bm.AllocatePage()
		data := "hello world!"
		// write data into page
		copy(bf.Page.data[:], data)
		assert.Equal(t, PID(0), bf.pid)

		// allocate another new page
		bf2 := bm.AllocatePage()
		assert.Equal(t, PID(1), bf2.pid)
	})

	t.Run("sync", func(t *testing.T) {
		bm := newBufferManager(-1, Partition4)
		wg := sync.WaitGroup{}
		pageCount := 12
		var (
			allocated BufferFrames
			mu        sync.Mutex
		)

		for i := 0; i < pageCount; i++ {
			wg.Add(1)
			go func() {
				b := bm.AllocatePage()

				mu.Lock()
				allocated = append(allocated, b)
				mu.Unlock()

				wg.Done()
			}()
		}

		wg.Wait()

		sort.Sort(allocated)
		AssertAllocatedPages(t, allocated, 4, 0, pageCount)
	})

}
