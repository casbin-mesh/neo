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
	"unsafe"
)

type Kind uint8

const (
	Leaf Kind = iota
	Node4
	Node16
	Node48
	Node256
)

var (
	NodeString = []string{"Leaf", "Node4", "Node16", "Node48", "Node256"}
)

func (k Kind) String() string {
	return NodeString[k]
}

// artNode included in all various size nodes
// node sizes: 8B + 3B + 2B(MaxPrefixLen)
type artNode struct {
	ref         unsafe.Pointer
	kind        Kind
	numChildren uint8
	partialLen  int
	partial     [MaxPrefixLen]byte
}

func (an artNode) isLeaf() bool {
	return an.kind == Leaf
}

func (an artNode) castLeft() *artLeaf {
	return (*artLeaf)(an.ref)
}

// node constraints
const (
	node4Max = 4

	node16Max = 16

	node48Max = 48

	node256Max = 256
)

type node4 struct {
	artNode
	children [node4Max]*artNode
	keys     [node4Max]byte
}

type node16 struct {
	artNode
	children [node16Max]*artNode
	keys     [node16Max]byte
}

type node48 struct {
	artNode
	children [node48Max]*artNode
	keys     [node256Max]byte
}

type node256 struct {
	artNode
	children [node256Max]*artNode
	keys     [node256Max]byte
}

func newNode4() *artNode {
	return &artNode{kind: Node4, ref: unsafe.Pointer(&node4{})}
}

func newNode16() *artNode {
	return &artNode{kind: Node16, ref: unsafe.Pointer(&node16{})}
}

func newNode48() *artNode {
	return &artNode{kind: Node48, ref: unsafe.Pointer(&node48{})}
}

func newNode256() *artNode {
	return &artNode{kind: Node256, ref: unsafe.Pointer(&node256{})}
}

func (an *artNode) node4() *node4 {
	return (*node4)(an.ref)
}

func (an *artNode) node16() *node16 {
	return (*node16)(an.ref)
}

func (an *artNode) node48() *node48 {
	return (*node48)(an.ref)
}

func (an *artNode) node256() *node256 {
	return (*node256)(an.ref)
}

func (an *artNode) leaf() *artLeaf {
	return (*artLeaf)(an.ref)
}

// Node helpers
func replaceRef(oldNode **artNode, newNode *artNode) {
	*oldNode = newNode
}

func replaceNode(oldNode *artNode, newNode *artNode) {
	*oldNode = *newNode
}

func (an *artNode) addChild256(char byte, child *artNode) bool {
	n := an.node256()
	n.numChildren++
	n.children[char] = child
	return false
}

func (an *artNode) addChild48(char byte, child *artNode) (grew bool) {
	n := an.node48()
	if n.numChildren < node48Max {
		index := byte(0)
		for n.children[index] != nil {
			index++
		}
		n.keys[char] = index + 1 // 0 means key is not exist
		n.children[index] = child
		n.numChildren++
	} else {
		newNode := an.grow()
		newNode.addChild(char, child)
		replaceNode(an, newNode)
		grew = true
	}
	return
}

func (an *artNode) addChild16(char byte, child *artNode) (grew bool) {
	n := an.node16()
	if n.numChildren < node16Max {
		idx := uint8(0)
		// find a slot
		for ; idx < n.numChildren; idx++ {
			if char < n.keys[idx] {
				break
			}
		}
		// shift
		copy(n.keys[idx+1:], n.keys[idx:])
		copy(n.children[idx+1:], n.children[idx:])
		// overwrite idx
		n.keys[idx] = char
		n.children[idx] = child
		n.numChildren++
	} else {
		newNode := an.grow()
		newNode.addChild(char, child)
		replaceNode(an, newNode)
		grew = true
	}

	return
}

// addChild4
func (an *artNode) addChild4(char byte, child *artNode) (grew bool) {
	n := an.node4()

	if n.numChildren < node4Max {
		idx := uint8(0)
		// find a slot
		for ; idx < n.numChildren; idx++ {
			if char < n.keys[idx] {
				break
			}
		}
		// shift
		copy(n.keys[idx+1:], n.keys[idx:])
		copy(n.children[idx+1:], n.children[idx:])
		// overwrite idx
		n.keys[idx] = char
		n.children[idx] = child
		n.numChildren++
	} else { // growing
		newNode := an.grow()
		newNode.addChild(char, child)
		replaceNode(an, newNode)
		grew = true
	}
	return
}

func (an *artNode) addChild(char byte, child *artNode) (grew bool) {
	switch an.kind {
	case Node4:
		return an.addChild4(char, child)
	case Node16:
		return an.addChild16(char, child)
	case Node48:
		return an.addChild48(char, child)
	case Node256:
		return an.addChild256(char, child)
	}
	return
}

func cloneMeta(dst *artNode, src *artNode) {
	if src == nil || dst == nil {
		return
	}
	dst.numChildren = src.numChildren
	dst.partial = src.partial
	dst.partialLen = src.partialLen
}

func (an *artNode) grow() *artNode {
	switch an.kind {
	case Node4:
		node := newNode16()
		cloneMeta(node, an)
		dst := node.node16()
		src := node.node4()
		copy(dst.keys[:], src.keys[:])
		copy(dst.children[:], src.children[:])
		return node
	case Node16:
		node := newNode48()
		cloneMeta(node, an)
		dst := node.node48()
		src := node.node16()
		copy(dst.keys[:], src.keys[:])
		copy(dst.children[:], src.children[:])
		return node
	case Node48:
		node := newNode256()
		cloneMeta(node, an)
		dst := node.node256()
		src := node.node48()
		for i := 0; i < node256Max; i++ {
			if index := src.keys[i]; index > 0 { // index = 0 means key is not exist
				dst.children[i] = src.children[index-1]
			}
		}
		return node
	}

	return nil
}

// checkPrefix Returns the number of prefix characters shared between
// the key and node.
func (an *artNode) checkPrefix(key Key, depth uint32) uint32 {
	maxCmp := min(min(MaxPrefixLen, an.partialLen), len(key)-int(depth))
	idx := uint32(0)
	for ; idx < uint32(maxCmp); idx++ {
		if an.partial[idx] != key[depth+idx] {
			return idx
		}
	}
	return idx
}

// prefixMismatch return the index at which the prefix mismatched
func (an *artNode) prefixMismatch(key Key, depth uint32) uint32 {
	maxCmp := min(min(MaxPrefixLen, an.partialLen), len(key)-int(depth))
	idx := uint32(0)
	for ; idx < uint32(maxCmp); idx++ {
		if an.partial[idx] != key[depth+idx] {
			return idx
		}
	}
	// check the leftmost(minimum) node
	if an.partialLen > MaxPrefixLen {
		l := leftmost(an)
		maxCmp = min(len(l.key), len(key))
		for ; idx < uint32(maxCmp); idx++ {
			if l.key[idx+depth] != key[depth+idx] {
				return idx
			}
		}
	}
	return idx
}

func (an *artNode) setPrefix(key Key, prefixLen int) {
	an.partialLen = prefixLen
	copy(an.partial[:], key[:min(prefixLen, MaxPrefixLen)])
}

func (an *artNode) findChildAndIdx(key byte) (*artNode, int) {
	idx := an.index(key)
	if idx != -1 {
		switch an.kind {
		case Node4:
			return an.node4().children[idx], idx

		case Node16:
			return an.node16().children[idx], idx

		case Node48:
			return an.node48().children[idx], int(key)
		case Node256:
			return an.node256().children[idx], int(key)
		}
	}
	return nil, -1
}

func (an *artNode) findChild(key byte) *artNode {
	idx := an.index(key)
	if idx != -1 {
		switch an.kind {
		case Node4:
			return an.node4().children[idx]

		case Node16:
			return an.node16().children[idx]

		case Node48:
			return an.node48().children[idx]
		case Node256:
			return an.node256().children[idx]
		}
	}
	return nil
}

func (an *artNode) index(char byte) int {
	switch an.kind {
	case Node4:
		n4 := an.node4()
		for idx := 0; idx < int(n4.numChildren); idx++ {
			if char == n4.keys[idx] {
				return idx
			}
		}
	case Node16:
		n16 := an.node16()
		for idx := 0; idx < int(n16.numChildren); idx++ {
			if char == n16.keys[idx] {
				return idx
			}
		}
	case Node48:
		n48 := an.node48()
		// for node48 keys, the 0 means not exists
		// in addChild, we shift 1 to store the child
		// now we need shift -1 to retrieve the child
		if idx := n48.keys[char]; idx > 0 {
			return int(idx) - 1
		}
	case Node256:
		return int(char)
	}
	return -1
}

func (an *artNode) removeChild256(char byte) uint8 {
	n := an.node256()
	n.children[char] = nil
	an.numChildren--
	return an.numChildren
}

func (an *artNode) removeChild48(char byte) uint8 {
	n := an.node48()
	pos := n.keys[char]
	n.keys[char] = 0
	n.children[pos-1] = nil

	an.numChildren--
	return an.numChildren
}

func (an *artNode) removeChild16At(idx int) uint8 {
	n := an.node16()
	copy(n.keys[idx:], n.keys[idx+1:])
	copy(n.children[idx:], n.children[idx+1:])

	an.numChildren--
	return an.numChildren
}

func (an *artNode) removeChild4At(idx int) uint8 {
	n := an.node4()
	copy(n.keys[idx:], n.keys[idx+1:])
	copy(n.children[idx:], n.children[idx+1:])

	an.numChildren--
	return an.numChildren
}

func (an *artNode) removeChildAt(idxOrChar byte) (shrank bool) {
	var (
		numChildren uint8
		minChildren uint16
	)
	switch an.kind {
	case Node4:
		numChildren = an.removeChild4At(int(idxOrChar))
		minChildren = 1
	case Node16:
		numChildren = an.removeChild16At(int(idxOrChar))
		minChildren = 3
	case Node48:
		numChildren = an.removeChild48(idxOrChar)
		minChildren = 12
	case Node256:
		numChildren = an.removeChild256(idxOrChar)
		minChildren = 37
	}

	if uint16(numChildren) == minChildren {
		newNode := an.shrink()
		replaceNode(an, newNode)
		return true
	}

	return false
}

func (an *artNode) shrink() *artNode {
	switch an.kind {
	case Node4:
		n := an.node4()
		child := n.children[0]
		if child.isLeaf() {
			return child
		}
		// concatenate the prefixes
		prefixLen := n.partialLen
		if prefixLen < MaxPrefixLen {
			n.partial[prefixLen] = n.keys[0]
			prefixLen++
		}
		if prefixLen < MaxPrefixLen {
			childPrefixLen := min(child.partialLen, MaxPrefixLen-prefixLen)
			// copy reset prefix
			copy(n.partial[prefixLen:], child.partial[:childPrefixLen])
			prefixLen += childPrefixLen
		}
		// store the prefix in child
		copy(child.partial[:], an.partial[:min(prefixLen, MaxPrefixLen)])
		child.partialLen += an.partialLen + 1

		return child
	case Node16:
		newNode := newNode4()
		cloneMeta(newNode, an)
		dst := newNode.node4()
		src := an.node16()
		copy(dst.keys[:], src.keys[:node4Max])
		copy(dst.children[:], src.children[:node4Max])
		return newNode
	case Node48:
		newNode := newNode16()
		cloneMeta(newNode, an)
		src := an.node48()
		dst := newNode.node16()
		child := 0
		for i := 0; i < node256Max; i++ {
			pos := src.keys[i]
			if pos > 0 {
				dst.keys[child] = byte(i)
				dst.children[child] = src.children[pos-1]
			}
		}
		return newNode
	case Node256:
		newNode := newNode48()
		cloneMeta(newNode, an)
		src := an.node256()
		dst := newNode.node48()
		for i := 0; i < node256Max; i++ {
			pos := src.keys[i]
			if pos > 0 {
				dst.children[pos] = src.children[i]
				dst.keys[i] = pos + 1
			}
		}
		return newNode
	}
	return nil
}
