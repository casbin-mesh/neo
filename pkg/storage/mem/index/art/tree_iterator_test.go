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
	"github.com/stretchr/testify/assert"
	"regexp"
	"sort"
	"testing"
)

func TestArtTree_Iterator(t *testing.T) {
	tree := NewArtTree()
	Seed(tree)
	sort.Sort(Data)

	var actual Sets

	for it := tree.Iterator(); it.HasNext(); {
		n, err := it.Next()
		actual.Append(n)
		assert.Nil(t, err)
	}

	assert.Equal(t, Data, actual)
}

func TestArtTree_IteratorNode48(t *testing.T) {
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
	for it := tree.Iterator(); it.HasNext(); {
		n, err := it.Next()
		actual.Append(n)
		assert.Nil(t, err)
	}
	assert.Equal(t, expected, actual)
}

func TestArtTree_Iterator2(t *testing.T) {
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
	for it := tree.Iterator(); it.HasNext(); {
		n, err := it.Next()
		actual.Append(n)
		assert.Nil(t, err)
	}
	assert.Equal(t, expected, actual)
}

func TestArtTree_IteratorEmpty(t *testing.T) {
	tree := NewArtTree()
	it := tree.Iterator(TraverseAll)
	_, err := it.Next()
	assert.Equal(t, ErrNoMoreNodes, err)
	assert.False(t, it.HasNext())
}

func TestArtTree_NewIteratorWithPrefix(t *testing.T) {
	lp := "program::1::info::common::name"
	tree := NewArtTree()
	Seed(tree)
	sort.Sort(Data)

	for i := 0; i < len(lp); i++ {
		regex := `^` + lp[:i+1]
		expected := Data.Filter(regexp.MustCompile(`^` + lp[:i+1]))

		var actual Sets
		for it := tree.NewIterator(NewIteratorConfig{
			prefix:  Key(lp[:i+1]),
			options: TraverseLeaf,
		}); it.HasNext(); {
			n, err := it.Next()
			actual.Append(n)
			assert.Nil(t, err)
		}
		assert.Equal(t, expected, actual, "failed at %s", regex)
	}
}
