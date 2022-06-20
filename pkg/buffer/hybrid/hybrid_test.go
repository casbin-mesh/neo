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
