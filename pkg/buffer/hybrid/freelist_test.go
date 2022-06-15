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
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		fl := FreeList{}
		assert.NotNil(t, fl)
	})
}

func TestFreeList_Push(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		fl := FreeList{}
		// alloc buffer frames
		l := 10
		data := make([]BufferFrame, l)
		for i := 0; i < l; i++ {
			fl.Push(&data[i])
		}

		// testing
		assert.Equal(t, uint64(l), fl.counter)
		cur := fl.head.Load().(*BufferFrame)
		assert.NotNil(t, cur)
		for i := l - 1; i >= 0; i-- {
			expected := &data[i]
			assert.Equal(t, expected, cur)
			cur = cur.nextFreeBF
		}
	})
	t.Run("sync", func(t *testing.T) {
		fl := FreeList{}
		// alloc buffer frames
		l := 10
		data := make([]BufferFrame, l)
		wg := sync.WaitGroup{}
		for i := 0; i < l; i++ {
			wg.Add(1)
			go func(bf *BufferFrame) {
				fl.Push(bf)
				wg.Done()
			}(&data[i])
		}
		wg.Wait()

		assert.Equal(t, uint64(l), fl.counter)
		cur := fl.head.Load().(*BufferFrame)
		count := 0
		for cur != nil {
			count++
			cur = cur.nextFreeBF
		}
		assert.Equal(t, l, count)
	})
}

func TestFreeList_Pop(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		fl := FreeList{}
		// alloc buffer frames
		l := 10
		data := make([]BufferFrame, l)
		for i := 0; i < l; i++ {
			fl.Push(&data[i])
		}

		for i := l - 1; i >= 0; i-- {
			got := fl.Pop()
			expected := data[fl.counter]
			assert.Equalf(t, expected, *got, "%d, it should be equal data[%d]", i, fl.counter)
		}
	})
	t.Run("sync", func(t *testing.T) {
		fl := FreeList{}
		// alloc buffer frames
		l := 10
		data := make([]BufferFrame, l)
		for i := 0; i < l; i++ {
			fl.Push(&data[i])
		}

		wg := sync.WaitGroup{}
		for i := 0; i < l; i++ {
			wg.Add(1)
			go func() {
				fl.Pop()
				wg.Done()
			}()
		}
		wg.Wait()

		assert.Equal(t, uint64(0), fl.counter)
	})
}
