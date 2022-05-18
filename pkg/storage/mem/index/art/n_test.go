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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewNode(t *testing.T) {
	n4 := newNode4()
	assert.NotNil(t, n4)
	assert.Equal(t, Node4, n4.kind)

	n16 := newNode16()
	assert.NotNil(t, n16)
	assert.Equal(t, Node16, n16.kind)

	n48 := newNode48()
	assert.NotNil(t, n48)
	assert.Equal(t, Node48, n48.kind)

	n256 := newNode256()
	assert.NotNil(t, n256)
	assert.Equal(t, Node256, n256.kind)

	l := newLeaf(Key("test"), Value("value1"))
	assert.Equal(t, Leaf, l.kind)

	assert.Equal(t, "Node4", n4.kind.String())
	assert.Equal(t, "Node16", n16.kind.String())
	assert.Equal(t, "Node48", n48.kind.String())
	assert.Equal(t, "Node256", n256.kind.String())
	assert.Equal(t, "Leaf", l.kind.String())
}

func TestLeaf_Basic(t *testing.T) {
	l := newLeaf(Key("test"), Value("value1"))

	assert.False(t, l.leaf().Match(Key("should be mismatched")))
	// we cannot shrink/grow leaf node
	assert.Nil(t, l.shrink())
	assert.Nil(t, l.grow())
}

func TestLeaf_PrefixMatch(t *testing.T) {
	l := newLeaf(Key("test"), Value("value1"))

	assert.False(t, l.leaf().Match(Key("should be mismatched")))
	assert.False(t, l.leaf().Match(nil))
	assert.True(t, l.leaf().Match(Key("test")))
}

func TestArtNode_SetPrefix(t *testing.T) {
	n4 := newNode4()
	n := n4.node()
	assert.NotNil(t, n)
	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n.setPrefix(key, 2)
	assert.True(t, n.partialLen == 2)
	assert.Equal(t, uint8(1), n.partial[0])
	assert.Equal(t, uint8(2), n.partial[1])

	n.setPrefix(key, MaxPrefixLen)
	assert.True(t, n.partialLen == MaxPrefixLen)
	assert.True(t, bytes.Compare(n.partial[:], key[:MaxPrefixLen]) == 0)
}

func TestArtNode_CheckPrefix(t *testing.T) {
	n := newNode4()
	n4 := n.node()
	assert.NotNil(t, n)
	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n4.setPrefix(key, len(key))

	assert.Equal(t, uint32(len(key)), n.checkPrefix(key, 0))

	// set a shorter key
	n4.setPrefix(key, 5)
	assert.Equal(t, uint32(5), n.checkPrefix(key, 0))
	assert.Equal(t, uint32(0), n.checkPrefix(key, 1))
	assert.Equal(t, uint32(5), n.checkPrefix(append([]byte{0}, key...), 1))
}

func TestArtNode_CloneMeta(t *testing.T) {
	// mock
	src := newNode4()
	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	src.node().setPrefix(key, len(key))
	src.node().numChildren = 255

	dst := newNode4()
	cloneMeta(dst.node(), src.node())

	assert.Equal(t, uint32(len(key)), dst.checkPrefix(key, 0))
	assert.Equal(t, uint8(255), dst.node().numChildren)
}

func TestArtNode_LeafFindChile(t *testing.T) {
	leaf := newLeaf(Key("I'm Key"), Value("I'm Value"))
	assert.Nil(t, leaf.findChild('I'))
}

type NodeTests struct {
	name        string
	target      *artNode
	maxChildren int
}

func TestArtTree_NodeAddChild(t *testing.T) {
	tests := []NodeTests{
		{"node 4", newNode4(), 4},
		{"node 16", newNode16(), 16},
		{"node 48", newNode48(), 48},
		{"node 256", newNode256(), 256},
	}

	for _, test := range tests {
		// insert leafs
		for i := 0; i < test.maxChildren; i++ {
			test.target.addChild(byte(i), newLeaf([]byte{byte(i)}, []byte{byte(i)}))
		}
		// check inserted items
		for i := 0; i < test.maxChildren; i++ {
			leaf := test.target.findChild(byte(i))
			assert.NotNilf(t, leaf, "should get a leaf %d in test %s", i, test.name)
			assert.Equalf(t, Value([]byte{byte(i)}), leaf.leaf().value, "should get a value equals %d in test %s", i, test.name)
		}
	}
}

func TestArtTree_NodeAddChildReverse(t *testing.T) {
	tests := []NodeTests{
		{"node 4", newNode4(), 4},
		{"node 16", newNode16(), 16},
		{"node 48", newNode48(), 48},
		{"node 256", newNode256(), 256},
	}

	for _, test := range tests {
		// insert leafs
		for i := test.maxChildren - 1; i >= 0; i-- {
			test.target.addChild(byte(i), newLeaf([]byte{byte(i)}, []byte{byte(i)}))
		}
	}

	for _, test := range tests {
		// check inserted items
		for i := 0; i < test.maxChildren; i++ {
			leaf := test.target.findChild(byte(i))
			assert.NotNilf(t, leaf, "should get a leaf %d in test %s", i, test.name)
			assert.Equalf(t, Value([]byte{byte(i)}), leaf.leaf().value, "should get a value equals %d in test %s", i, test.name)
		}
	}
}

func TestArtLeaf_LeafAddChild(t *testing.T) {
	leaf := newLeaf(Key("I'm Key"), Value("I'm Value"))
	assert.Falsef(t, leaf.addChild(byte('b'), nil), "should cannot add child on leaf nodes")
}

func TestArtTree_NodeIndex(t *testing.T) {
	tests := []NodeTests{
		{"node 4", newNode4(), 4},
		{"node 16", newNode16(), 16},
		{"node 48", newNode48(), 48},
		{"node 256", newNode256(), 256},
	}

	for _, test := range tests {
		// insert leafs
		for i := 0; i < test.maxChildren; i++ {
			test.target.addChild(byte(i), newLeaf([]byte{byte(i)}, []byte{byte(i)}))
		}
		// check inserted items
		for i := 0; i < test.maxChildren; i++ {
			assert.Equal(t, i, test.target.index(byte(i)))
		}
	}
}

func TestArtTree_NodeMinimumMaximum(t *testing.T) {
	tests := []NodeTests{
		{"node 4", newNode4(), 4},
		{"node 16", newNode16(), 16},
		{"node 48", newNode48(), 48},
		{"node 256", newNode256(), 256},
	}

	for _, test := range tests {
		// insert leafs
		for i := 0; i < test.maxChildren; i++ {
			test.target.addChild(byte(i), newLeaf([]byte{byte(i)}, []byte{byte(i)}))
		}
	}

	for _, test := range tests {
		// check minimum and maximum
		minimum := leftmost(test.target)

		assert.Equalf(t, Key([]byte{0}), minimum.key, "should equals 0 in test %s", test.name)
		assert.Equalf(t, Value([]byte{0}), minimum.value, "should equals 0 in test %s", test.name)

		maximum := rightmost(test.target)
		assert.Equalf(t, Key([]byte{byte(test.maxChildren - 1)}), maximum.key, "should equals %d in test %s", test.maxChildren-1, test.name)
		assert.Equalf(t, Value([]byte{byte(test.maxChildren - 1)}), maximum.value, "should equals %d in test %s", test.maxChildren-1, test.name)
	}
}

func TestArtTree_NodeGrow(t *testing.T) {
	nodes := []*artNode{newNode4(), newNode16(), newNode48()}
	expected := []Kind{Node16, Node48, Node256}

	for i, node := range nodes {
		newNode := node.grow()
		assert.Equal(t, expected[i], newNode.kind)
	}
}

func TestArtTree_NodeShrink(t *testing.T) {
	nodes := []*artNode{newNode4(), newNode16(), newNode48(), newNode256()}
	expected := []Kind{Leaf, Node4, Node16, Node48}

	for i, node := range nodes {
		// add some nodes
		for j := 0; j < 3; j++ {
			node.addChild(byte(j), newLeaf(Key{byte(i)}, Value("value")))
		}
		newNode := node.shrink()
		assert.Equal(t, expected[i], newNode.kind)
	}
}
