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

package lru

import (
	"github.com/casbin-mesh/neo/pkg/buffer/replacer"
	"github.com/stretchr/testify/assert"
	"sort"
	"sync"
	"testing"
)

func TestSimple(t *testing.T) {
	r := NewLRUReplacer(5)
	// unpin 5 frame
	for i := 0; i < 3; i++ {
		assert.Nil(t, r.Unpin(uint64(i)))
	}
	for i := 0; i < 3; i++ {
		assert.Nil(t, r.Unpin(uint64(i)))
	}
	assert.Equal(t, uint64(3), r.Size())

	// victim 5 frame
	for i := 0; i < 3; i++ {
		var res uint64
		assert.True(t, r.Victim(&res))
		assert.Equal(t, res, uint64(i))
	}

	// pin 5 frame
	for i := 0; i < 5; i++ {
		assert.Nil(t, r.Pin(uint64(i)))
	}
	assert.Equal(t, uint64(0), r.Size())

	// should be unable to victim
	var res uint64
	assert.False(t, r.Victim(&res))
}

func UnpinHelper(r replacer.Replacer, frameId uint64, t assert.TestingT) {
	assert.Nil(t, r.Unpin(frameId))
}

func TestConcurrencyUnpin(t *testing.T) {
	size := 1000
	r := NewLRUReplacer(uint64(size))
	wg := sync.WaitGroup{}
	for i := 0; i < size; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			UnpinHelper(r, uint64(num), t)
		}(i)
	}
	wg.Wait()
}

func PinHelper(r replacer.Replacer, frameId uint64, t assert.TestingT) {
	assert.Nil(t, r.Pin(frameId))
}

func TestConcurrencyPin(t *testing.T) {
	size := 1000
	r := NewLRUReplacer(uint64(size))
	wg := sync.WaitGroup{}
	for i := 0; i < size; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			PinHelper(r, uint64(num), t)
		}(i)
	}
	wg.Wait()
}

type Task func(r replacer.Replacer, frameId uint64, t assert.TestingT)

func TestConcurrencyPinAndUnPin(t *testing.T) {
	var tasks []Task
	size := uint64(1024)
	r := NewLRUReplacer(size)
	// generate test tasks
	for i := uint64(0); i < size; i++ {
		if i%2 == 0 {
			tasks = append(tasks, PinHelper)
		} else {
			tasks = append(tasks, UnpinHelper)
		}
	}

	wg := sync.WaitGroup{}
	for i, task := range tasks {
		wg.Add(1)
		go func(num int, ts Task) {
			defer wg.Done()
			ts(r, uint64(num), t)
		}(i, task)
	}

	wg.Wait()
}

func TestConcurrencyUnpinAndVictim(t *testing.T) {
	var tasks []Task
	size := uint64(1024)
	r := NewLRUReplacer(size)
	// generate test tasks
	for i := uint64(0); i < size; i++ {
		tasks = append(tasks, UnpinHelper)
	}

	wg := sync.WaitGroup{}
	for i, task := range tasks {
		wg.Add(1)
		go func(num int, ts Task) {
			defer wg.Done()
			ts(r, uint64(num), t)
		}(i, task)
	}
	wg.Wait()

	var res uint64
	var expected []uint64
	var results []uint64
	for i := uint64(0); i < size; i++ {
		assert.True(t, r.Victim(&res))
		expected = append(expected, i)
		results = append(results, res)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	assert.Equal(t, expected, results)
}

func BenchmarkLruReplacer_Victim(b *testing.B) {
	size := 1000000
	r := NewLRUReplacer(uint64(size))
	// warm up
	for i := 0; i < size; i++ {
		assert.Nil(b, r.Unpin(uint64(i)))
	}
	b.ResetTimer()
	var res uint64
	r.Victim(&res)
}

func BenchmarkLruReplacer_Pin(b *testing.B) {
	size := 1000000
	r := NewLRUReplacer(uint64(size))
	// warm up
	for i := 0; i < size; i++ {
		assert.Nil(b, r.Unpin(uint64(i)))
	}
	b.ResetTimer()
	assert.Nil(b, r.Pin(uint64(size+1)))
}

func BenchmarkLruReplacer_Unpin(b *testing.B) {
	size := 1000000
	r := NewLRUReplacer(uint64(size))
	// warm up
	for i := 0; i < size-1; i++ {
		assert.Nil(b, r.Unpin(uint64(i)))
	}
	b.ResetTimer()
	r.Pin(uint64(size))
}

func TestLruReplacer_Victim(t *testing.T) {
	size := uint64(1024)
	r := NewLRUReplacer(size)
	generateVictimTest(r, size, t)
}

func TestLruReplacer_Pin(t *testing.T) {
	size := uint64(1024)
	r := NewLRUReplacer(size)
	generatePinTest(r, size, t)
}

func TestLruReplacer_Size(t *testing.T) {
	size := uint64(1024)
	r := NewLRUReplacer(size)
	generateSizeTest(r, size, t)
}
