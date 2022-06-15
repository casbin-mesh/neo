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

// origin https://github.com/tidwall/btree-benchmark/blob/main/main.go

package art

import (
	"bytes"
	"encoding/binary"
	"github.com/casbin-mesh/neo/pkg/storage/mem/index/art"
	"github.com/dgraph-io/badger/v3/skl"
	"github.com/dgraph-io/badger/v3/y"
	sArt "github.com/dshulyak/art"
	gbtree "github.com/google/btree"
	tbtree "github.com/tidwall/btree"

	"github.com/tidwall/lotsa"
	"math/rand"
	"os"
	"sort"
	"testing"
)

type item struct {
	key []byte
}

func (i item) Less(other gbtree.Item) bool {
	return bytes.Compare(i.key, other.(item).key) < 0
}

func lessG(a, b item) bool {
	return bytes.Compare(a.key, b.key) < 0
}

func less(a, b interface{}) bool {
	return bytes.Compare(a.(item).key, b.(item).key) < 0
}

func newBTree() *tbtree.BTree {
	return tbtree.NewNonConcurrent(less)
}

func newBTreeG() *tbtree.Generic[item] {
	return tbtree.NewGenericOptions[item](lessG, tbtree.Options{NoLocks: true})
}

func Test_Bench(t *testing.T) {
	N := 1_000_000
	keys := make([]item, N)
	for i := 0; i < N; i++ {
		var buf [8]byte
		binary.PutVarint(buf[:], int64(i))
		keys[i] = item{key: buf[:]}
	}
	lotsa.Output = os.Stdout
	lotsa.MemUsage = true

	sortInts := func() {
		sort.Slice(keys, func(i, j int) bool {
			return less(keys[i], keys[j])
		})
	}

	shuffleInts := func() {
		for i := range keys {
			j := rand.Intn(i + 1)
			keys[i], keys[j] = keys[j], keys[i]
		}
	}

	degree := 128

	println()
	println("** sequential set **")
	sortInts()

	// skl
	print("sklist:     set-seq        ")
	sk := skl.NewSkiplist(int64((N + 1) * skl.MaxNodeSize))
	lotsa.Ops(N, 1, func(i, _ int) {
		sk.Put(keys[i].key, y.ValueStruct{Value: nil, Meta: 0, UserMeta: 0})
	})
	// sartTree
	print("sArt:      set-seq        ")
	sart := sArt.Tree{}
	lotsa.Ops(N, 1, func(i, _ int) {
		sart.Insert(keys[i].key, nil)
	})
	// artTree
	print("artTree:    set-seq        ")
	artTree := art.NewArtTree()
	lotsa.Ops(N, 1, func(i, _ int) {
		artTree.Insert(keys[i].key, nil)
	})

	// google
	print("google:     set-seq        ")
	gtr := gbtree.New(degree)
	lotsa.Ops(N, 1, func(i, _ int) {
		gtr.ReplaceOrInsert(keys[i])
	})

	// non-generics tidwall
	print("tidwall:    set-seq        ")
	ttr := newBTree()
	lotsa.Ops(N, 1, func(i, _ int) {
		ttr.Set(keys[i])
	})
	print("tidwall(G): set-seq        ")
	ttrG := newBTreeG()
	lotsa.Ops(N, 1, func(i, _ int) {
		ttrG.Set(keys[i])
	})

	print("tidwall:    set-seq-hint   ")
	ttr = newBTree()
	var hint tbtree.PathHint
	lotsa.Ops(N, 1, func(i, _ int) {
		ttr.SetHint(keys[i], &hint)
	})
	print("tidwall(G): set-seq-hint   ")
	ttrG = newBTreeG()
	var hintG tbtree.PathHint
	lotsa.Ops(N, 1, func(i, _ int) {
		ttrG.SetHint(keys[i], &hintG)
	})
	print("tidwall:    load-seq       ")
	ttr = newBTree()
	lotsa.Ops(N, 1, func(i, _ int) {
		ttr.Load(keys[i])
	})
	print("tidwall(G): load-seq       ")
	ttrG = newBTreeG()
	lotsa.Ops(N, 1, func(i, _ int) {
		ttrG.Load(keys[i])
	})

	// go array
	print("go-arr:     append         ")
	var arr []item
	lotsa.Ops(N, 1, func(i, _ int) {
		arr = append(arr, keys[i])
	})

	println()
	println("** sequential get **")
	sortInts()
	print("sklist:     get-seq        ")
	lotsa.Ops(N, 1, func(i, _ int) {
		v := sk.Get(keys[i].key)
		if v.Value == nil {
			panic("not found")
		}
	})
	print("artTree:    get-seq        ")
	lotsa.Ops(N, 1, func(i, _ int) {
		_, found := artTree.Search(keys[i].key)
		if !found {
			panic("not found")
		}
	})
	print("sArt:      set-seq        ")
	lotsa.Ops(N, 1, func(i, _ int) {
		_, found := sart.Get(keys[i].key)
		if !found {
			panic("not found")
		}
	})
	print("google:     get-seq        ")
	lotsa.Ops(N, 1, func(i, _ int) {
		re := gtr.Get(keys[i])
		if re == nil {
			panic(re)
		}
	})
	print("tidwall:    get-seq        ")
	lotsa.Ops(N, 1, func(i, _ int) {
		re := ttr.Get(keys[i])
		if re == nil {
			panic(re)
		}
	})
	print("tidwall(G): get-seq        ")
	lotsa.Ops(N, 1, func(i, _ int) {
		re, ok := ttrG.Get(keys[i])
		if !ok {
			panic(re)
		}
	})
	print("tidwall:    get-seq-hint   ")
	lotsa.Ops(N, 1, func(i, _ int) {
		re := ttr.GetHint(keys[i], &hint)
		if re == nil {
			panic(re)
		}
	})
	print("tidwall(G): get-seq-hint   ")
	lotsa.Ops(N, 1, func(i, _ int) {
		re, ok := ttrG.GetHint(keys[i], &hintG)
		if !ok {
			panic(re)
		}
	})

	println()
	println("** random set **")
	shuffleInts()

	sk = skl.NewSkiplist(int64((N + 1) * skl.MaxNodeSize))
	print("sklist:    set-rand        ")
	lotsa.Ops(N, 1, func(i, _ int) {
		sk.Put(keys[i].key, y.ValueStruct{})
	})
	sart = sArt.Tree{}
	print("sArt:      set-rand        ")
	lotsa.Ops(N, 1, func(i, _ int) {
		sart.Insert(keys[i].key, nil)
	})
	print("artTree:    set-rand       ")
	artTree = art.NewArtTree()
	lotsa.Ops(N, 1, func(i, _ int) {
		artTree.Insert(keys[i].key, nil)
	})
	print("google:     set-rand       ")
	gtr = gbtree.New(degree)
	lotsa.Ops(N, 1, func(i, _ int) {
		gtr.ReplaceOrInsert(keys[i])
	})
	print("tidwall:    set-rand       ")
	ttr = newBTree()
	lotsa.Ops(N, 1, func(i, _ int) {
		ttr.Set(keys[i])
	})
	print("tidwall(G): set-rand       ")
	ttrG = newBTreeG()
	lotsa.Ops(N, 1, func(i, _ int) {
		ttrG.Set(keys[i])
	})
	print("tidwall:    set-rand-hint  ")
	ttr = newBTree()
	lotsa.Ops(N, 1, func(i, _ int) {
		ttr.SetHint(keys[i], &hint)
	})
	print("tidwall(G): set-rand-hint  ")
	ttrG = newBTreeG()
	lotsa.Ops(N, 1, func(i, _ int) {
		ttrG.SetHint(keys[i], &hintG)
	})
	print("tidwall:    set-after-copy ")
	ttr = ttr.Copy()
	lotsa.Ops(N, 1, func(i, _ int) {
		ttr.Set(keys[i])
	})
	print("tidwall(G): set-after-copy ")
	ttrG = ttrG.Copy()
	lotsa.Ops(N, 1, func(i, _ int) {
		ttrG.Set(keys[i])
	})
	print("tidwall:    load-rand      ")
	ttr = newBTree()
	lotsa.Ops(N, 1, func(i, _ int) {
		ttr.Load(keys[i])
	})
	print("tidwall(G): load-rand      ")
	ttrG = newBTreeG()
	lotsa.Ops(N, 1, func(i, _ int) {
		ttrG.Load(keys[i])
	})
	println()
	println("** random get **")

	shuffleInts()
	gtr = gbtree.New(degree)
	ttr = newBTree()
	ttrG = newBTreeG()
	for _, key := range keys {
		gtr.ReplaceOrInsert(key)
		ttrG.Set(key)
		ttr.Set(key)
	}
	shuffleInts()

	print("sklist:    get-rand        ")
	lotsa.Ops(N, 1, func(i, _ int) {
		v := sk.Get(keys[i].key)
		if v.Value == nil {
			panic("not found")
		}
	})
	print("sArt:      get-rand        ")
	lotsa.Ops(N, 1, func(i, _ int) {
		_, found := sart.Get(keys[i].key)
		if !found {
			panic("not found")
		}
	})
	print("artTree:    get-rand       ")
	lotsa.Ops(N, 1, func(i, _ int) {
		_, found := artTree.Search(keys[i].key)
		if !found {
			panic("not found")
		}
	})

	print("google:     get-rand       ")
	lotsa.Ops(N, 1, func(i, _ int) {
		re := gtr.Get(keys[i])
		if re == nil {
			panic(re)
		}
	})
	print("tidwall:    get-rand       ")
	lotsa.Ops(N, 1, func(i, _ int) {
		re := ttr.Get(keys[i])
		if re == nil {
			panic(re)
		}
	})
	print("tidwall(G): get-rand       ")
	lotsa.Ops(N, 1, func(i, _ int) {
		re, ok := ttrG.Get(keys[i])
		if !ok {
			panic(re)
		}
	})
	print("tidwall:    get-rand-hint  ")
	lotsa.Ops(N, 1, func(i, _ int) {
		re := ttr.GetHint(keys[i], &hint)
		if re == nil {
			panic(re)
		}
	})
	print("tidwall(G): get-rand-hint  ")
	lotsa.Ops(N, 1, func(i, _ int) {
		re, ok := ttrG.GetHint(keys[i], &hintG)
		if !ok {
			panic(re)
		}
	})

	println()
	println("** range **")
	print("sklist:     iter       ")
	lotsa.Ops(N, 1, func(i, _ int) {
		if i == 0 {
			it := sk.NewIterator()
			for ; it.Valid(); it.Next() {
			}
			it.Close()
		}
	})

	print("sArt:      iter        ")
	lotsa.Ops(N, 1, func(i, _ int) {
		if i == 0 {
			it := sart.Iterator(nil, nil)
			for it.Next() {
			}
		}
	})
	print("artTree:    traverse      ")
	lotsa.Ops(N, 1, func(i, _ int) {
		if i == 0 {
			artTree.Traversal(func(node *art.Node) bool {
				return true
			})
		}
	})
	print("artTree:    iter          ")
	lotsa.Ops(N, 1, func(i, _ int) {
		if i == 0 {
			for it := artTree.Iterator(); it.HasNext(); {
				it.Next()
			}
		}
	})
	print("google:     ascend        ")
	lotsa.Ops(N, 1, func(i, _ int) {
		if i == 0 {
			gtr.Ascend(func(item gbtree.Item) bool {
				return true
			})
		}
	})
	print("tidwall:    ascend        ")
	lotsa.Ops(N, 1, func(i, _ int) {
		if i == 0 {
			ttr.Ascend(nil, func(item interface{}) bool {
				return true
			})
		}
	})
	print("tidwall(G): iter          ")
	lotsa.Ops(N, 1, func(i, _ int) {
		if i == 0 {
			iter := ttrG.Iter()
			for ok := iter.First(); ok; ok = iter.Next() {
			}
			iter.Release()
		}
	})
	print("tidwall(G): scan          ")
	lotsa.Ops(N, 1, func(i, _ int) {
		if i == 0 {
			ttrG.Scan(func(item item) bool {
				return true
			})
		}
	})
	print("tidwall(G): walk          ")
	lotsa.Ops(N, 1, func(i, _ int) {
		if i == 0 {
			ttrG.Walk(func(items []item) bool {
				for j := 0; j < len(items); j++ {
				}
				return true
			})
		}
	})

	print("go-arr:     for-loop      ")
	lotsa.Ops(N, 1, func(i, _ int) {
		if i == 0 {
			for j := 0; j < len(arr); j++ {
			}
		}
	})

}
