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

package art

import (
	"bytes"
	"fmt"
	"github.com/dshulyak/art"
	"github.com/stretchr/testify/assert"
	tbtree "github.com/tidwall/btree"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestArtTest_Insert(t *testing.T) {
	tree := Tree[int]{}
	for i := 0; i < 1_000_000; i++ {
		tree.Insert(Key(fmt.Sprintf("sharedNode::%d", i)), i)
	}
}

func TestTree_ConcurrentInsert(t *testing.T) {
	t.Parallel()
	// set up
	N := 1_000_000
	tree := Tree[int]{}
	wg := sync.WaitGroup{}
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			tree.Insert(Key(fmt.Sprintf("sharedNode::%d", i)), i)
			wg.Done()
		}(i)
	}
	wg.Wait()
	for i := 0; i < N; i++ {
		value, found := tree.Search(Key(fmt.Sprintf("sharedNode::%d", i)))
		assert.True(t, found)
		assert.Equal(t, i, value)
	}
}

func TestTree_ConcurrentInsert2(t *testing.T) {
	t.Parallel()
	// set up
	N := 1_000_000
	tree := Tree[[]byte]{}
	inserted := []Key{}
	mu := sync.RWMutex{}
	wg := sync.WaitGroup{}

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			k := randomKey(rng)
			tree.Insert(k, k)
			mu.Lock()
			inserted = append(inserted, k)
			mu.Unlock()
			wg.Done()
		}(i)
	}
	wg.Wait()
	for _, key := range inserted {
		value, found := tree.Search(key)
		assert.True(t, found)
		assert.Equal(t, []byte(key), value)
	}
}

func BenchmarkArtConcurrentInsert(b *testing.B) {
	value := newValue(123)
	l := Tree[[]byte]{}
	b.ResetTimer()
	//var count int
	b.RunParallel(func(pb *testing.PB) {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			l.Insert(randomKey(rng), value)
		}
	})
}

func BenchmarkAnotherArtConcurrentInsert(b *testing.B) {
	value := newValue(123)
	l := art.Tree{}
	b.ResetTimer()
	//var count int
	b.RunParallel(func(pb *testing.PB) {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			l.Insert(randomKey(rng), value)
		}
	})
}

//func BenchmarkConcurrentInsert(b *testing.B) {
//	value := newValue(123)
//	l := NewArtTree()
//	b.ResetTimer()
//	//var count int
//	b.RunParallel(func(pb *testing.PB) {
//		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
//		for pb.Next() {
//			l.Insert(randomKey(rng), value)
//		}
//	})
//}

func BenchmarkBtreeConcurrentInsert(b *testing.B) {
	l := tbtree.NewGenericOptions[[]byte](func(a, b []byte) bool {
		return bytes.Compare(a, b) < 0
	}, tbtree.Options{NoLocks: false})
	b.ResetTimer()
	//var count int
	b.RunParallel(func(pb *testing.PB) {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			l.Set(randomKey(rng))
		}
	})
}
