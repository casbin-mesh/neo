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
	"github.com/stretchr/testify/assert"
	"regexp"
	"sort"
	"testing"
)

type Sets []Set

func (sets Sets) Len() int {
	return len(sets)
}

func (sets Sets) Less(i, j int) bool {
	return bytes.Compare(sets[i].key, sets[j].key) < 0
}

func (sets Sets) Swap(i, j int) {
	sets[i], sets[j] = sets[j], sets[i]
}

func (sets Sets) Filter(regexp *regexp.Regexp) Sets {
	var result Sets
	for _, set := range sets {
		if regexp.Match(set.key) {
			result = append(result, set)
		}
	}
	return result
}

var Data = Sets{
	// program
	{Key("program::1::info::common::name"), Value("weny")},
	{Key("program::1::info::common::version"), Value("v0.1.0")},
	{Key("program::1::info::common::dataset"), Value("dafaf")},
	{Key("program::1::info::address::code"), Value("22323234")},
	{Key("program::2::info::common::name"), Value("weny")},
	{Key("program::2::info::common::version"), Value("v0.1.0")},
	{Key("program::3::info::common::dataset"), Value("dafaf")},
	{Key("program::4::info::address::code"), Value("22323234")},
	{Key("program::5::info::common::name"), Value("weny")},
	{Key("program::6::info::common::version"), Value("v0.1.0")},
	{Key("program::7::info::common::dataset"), Value("dafaf")},
	{Key("program::10::info::address::code"), Value("22323234")},
	// register
	{Key("register::1::info::common::name"), Value("weny")},
	{Key("register::1::info::common::version"), Value("v0.1.0")},
	{Key("register::1::info::common::dataset"), Value("dafaf")},
	{Key("register::1::info::address::code"), Value("22323234")},
	{Key("register::2::info::common::name"), Value("weny")},
	{Key("register::2::info::common::version"), Value("v0.1.0")},
	{Key("register::3::info::common::dataset"), Value("dafaf")},
	{Key("register::4::info::address::code"), Value("22323234")},
	{Key("register::5::info::common::name"), Value("weny")},
	{Key("register::6::info::common::version"), Value("v0.1.0")},
	{Key("register::7::info::common::dataset"), Value("dafaf")},
	{Key("register::10::info::address::code"), Value("22323234")},
}

func Seed(tree *Tree) {
	for _, datum := range Data {
		tree.Insert(datum.key, datum.value)
	}
}

func (sets Sets) Print() {
	for _, set := range sets {
		fmt.Printf("%s ---> %s\n", set.key, set.value)
	}
}

func TestSets(t *testing.T) {
	sort.Sort(Data)
	sub := Data.Filter(regexp.MustCompile("^program::1"))
	for _, set := range sub {
		fmt.Printf("%s ---> %s\n", set.key, set.value)
	}
}

func TestArtTree_SeekPrefix(t *testing.T) {
	lp := "program::1::info::common::name"
	tree := NewArtTree()
	Seed(tree)
	sort.Sort(Data)

	for i := 0; i < len(lp); i++ {
		regex := `^` + lp[:i+1]
		expected := Data.Filter(regexp.MustCompile(`^` + lp[:i+1]))
		//fmt.Println(regex)
		//expected.Print()
		var actual Sets
		tree.Seek(Key(lp[:i+1]), AppendToSet(&actual))
		//actual.Print()
		assert.Equal(t, expected, actual, "failed at %s", regex)
	}
}

func TestArtTree_SeekEmptyPrefix(t *testing.T) {
	tree := NewArtTree()
	Seed(tree)
	sort.Sort(Data)

	var actual Sets
	tree.Seek(nil, AppendToSet(&actual))
	assert.Equal(t, Data, actual)
}

func TestArtTree_Traversal(t *testing.T) {
	tree := NewArtTree()
	Seed(tree)
	sort.Sort(Data)

	var actual Sets
	tree.Traversal(AppendToSet(&actual), TraverseAll)
	assert.Equal(t, Data, actual)
}

func AppendToSet(sets *Sets) Callback {
	return func(node *Node) bool {
		if node.IsLeaf() {
			leaf := node.Leaf()
			*sets = append(*sets, Set{
				key:   leaf.key.Clone(),
				value: leaf.value.Clone(),
			})
		}
		return true
	}
}

func (sets *Sets) Append(node *Node) {
	if node == nil {
		return
	}
	if node.IsLeaf() {
		leaf := node.Leaf()
		*sets = append(*sets, Set{
			key:   leaf.key.Clone(),
			value: leaf.value.Clone(),
		})
	}
}

func LeafPrint(node *Node) bool {
	if node == nil {
		return false
	}
	if node.IsLeaf() {
		leaf := node.Leaf()
		fmt.Printf("%s ---> %s\n", leaf.key, leaf.value)
	} else {
		//n := node.node()
		//fmt.Printf("inner node %v |%s| \n", n.partial, n.partial)
		//switch node.kind {
		//case Node4:
		//	for i, key := range node.node4().keys {
		//		//fmt.Printf("[%d]:%c\n", i, key)
		//	}
		//}
	}
	return true
}

func TestArtTree_Traversal2(t *testing.T) {
	g := NewKeyValueGenerator()
	tree := NewArtTree()
	var expected Sets
	for i := 0; i < 10000; i++ {
		k, v := g.next()
		tree.Insert(k, v)
		expected = append(expected, Set{
			key:   k.Clone(),
			value: v.Clone(),
		})
	}
	sort.Sort(expected)
	var actual Sets
	tree.Traversal(AppendToSet(&actual))
	assert.Equal(t, expected, actual)
}

func TestArtTree_TraversalNode48(t *testing.T) {
	g := NewKeyValueGenerator()
	tree := NewArtTree()
	var expected Sets
	for i := 0; i < 48; i++ {
		k, v := g.next()
		tree.Insert(k, v)
		expected = append(expected, Set{
			key:   k.Clone(),
			value: v.Clone(),
		})
	}
	sort.Sort(expected)
	var actual Sets
	tree.Traversal(AppendToSet(&actual))
	assert.Equal(t, expected, actual)
}

func TestArtTree_TraversalEmptyTree(t *testing.T) {
	tree := NewArtTree()
	var actual Sets
	tree.Traversal(AppendToSet(&actual))
	assert.Equal(t, 0, len(actual))
}
