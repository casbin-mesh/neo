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
	"sync"
	"unsafe"
)

const (
	PageSize        = 16 * 1024
	PageBodySize    = PageSize - 24
	BufferFrameSize = uint64(unsafe.Sizeof(BufferFrame{}))
	FREE            = State(iota)
	HOT
	COOL
	LOADED
)

type (
	State uint8
	Page  struct {
		// headers
		GSN        uint64 //global serial number
		dType      uint64 // Data Structure Type ID
		magicDebug uint64 // for debugging
		// data
		data [PageBodySize]byte
	}
	Header struct {
		pid            PID
		lastWrittenGSN uint64
		state          State
		keepInMemory   bool
		latch          sync.RWMutex
		nextFreeBF     *BufferFrame

		debug uint64
	} // in memory
	BufferFrame struct {
		Header
		Page
	}
)

func (bf *BufferFrame) reset() {
	bf.lastWrittenGSN = 0
	bf.state = FREE
	bf.nextFreeBF = nil
}

func (bf *BufferFrame) isFree() bool {
	return false
}

func (bf *BufferFrame) isDirty() bool {
	return false
}
