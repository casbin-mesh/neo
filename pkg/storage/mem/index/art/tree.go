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

type artTree struct {
	root *artNode
	size int
}

func NewArtTree() *artTree {
	return &artTree{
		root: nil,
		size: 0,
	}
}

func (art artTree) Size() int {
	return art.size
}

type Ordered interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64
}

func min[T Ordered](a T, b T) T {
	if a < b {
		return a
	}
	return b
}

// Find the leftmost(by default, minimum) leaf under a artNode
func leftmost(an *artNode) *artLeaf {
	switch an.kind {
	case Leaf:
		return an.leaf()

	case Node4:
		node := an.node4()
		if node.children[0] != nil {
			return leftmost(node.children[0])
		}

	case Node16:
		node := an.node16()
		if node.children[0] != nil {
			return leftmost(node.children[0])
		}

	case Node48:
		node := an.node48()
		idx := uint8(0)
		for ; node.children[node.keys[idx]] == nil; idx++ {
		}
		return leftmost(node.children[node.keys[idx]-1])

	case Node256:
		node := an.node256()
		idx := 0
		for ; node.children[idx] == nil; idx++ {
		}
		return leftmost(node.children[idx])

	}

	return nil // that should never happen in normal case
}

// Find the rightmost(by default, maximum) leaf under a artNode
func rightmost(an *artNode) *artLeaf {
	switch an.kind {
	case Leaf:
		return an.leaf()

	case Node4:
		node := an.node4()
		if node.children[node.numChildren-1] != nil {
			return rightmost(node.children[node.numChildren-1])
		}

	case Node16:
		node := an.node16()
		if node.children[node.numChildren-1] != nil {
			return rightmost(node.children[node.numChildren-1])
		}

	case Node48:
		node := an.node48()
		idx := uint8(255)
		for ; node.keys[idx] == 0; idx-- { // found the right most key
		}

		return rightmost(node.children[node.keys[idx]-1])

	case Node256:
		node := an.node256()
		idx := 255
		for ; node.children[idx] == nil; idx-- {
		}
		return rightmost(node.children[idx])
	}

	return nil // that should never happen in normal case
}

func (art *artTree) Insert(key Key, value Value) (Value, bool) {
	old, updated := art.recursiveInsert(&art.root, key, value, 0)
	if !updated {
		art.size++
	}
	return old, updated
}

func (art *artTree) Remove(key Key) (Value, bool) {
	value, deleted := art.recursiveRemove(&art.root, key, 0)
	if deleted {
		art.size--
		return value, true
	}
	return nil, false
}

func (art *artTree) Search(key Key) (Value, bool) {
	current := art.root
	depth := uint32(0)
	for current != nil {
		if current.isLeaf() {
			l := current.leaf()
			if l.Match(key) {
				return l.value, true
			}
			return nil, false
		}
		n := current.node()
		if n.partialLen > 0 {
			prefixLen := current.checkPrefix(key, depth)
			if prefixLen != uint32(min(n.partialLen, MaxPrefixLen)) {
				return nil, false
			}
			depth += uint32(n.partialLen)
		}

		next := current.findChild(key.At(int(depth)))
		if *next != nil {
			current = *next
		} else {
			current = nil
		}
		depth++

	}
	return nil, false
}

func (art *artTree) recursiveInsert(curNode **artNode, key Key, value Value, depth uint32) (Value, bool) {
	current := *curNode
	// if current is nil, insert a left node
	if current == nil {
		replaceRef(curNode, newLeaf(key, value))
		return value, false
	}

	// if current is a left node, we need replace it with a node (growing)
	if current.isLeaf() {
		left := current.castLeft()
		// update an existing value
		if left.Match(key) {
			oldValue := left.value
			left.value = value
			return oldValue, true
		}
		// new leaf
		newLeftNode := newLeaf(key, value)
		left2 := newLeftNode.leaf()
		longestPrefix := longestCommonPrefix(left, left2, depth)
		newNode := newNode4()
		n4 := newNode.node()
		if longestPrefix > 0 {
			n4.setPrefix(key[depth:], longestPrefix)
		}
		// it's safe to call add child directly
		newNode.addChild4(left.key.At(int(depth)+longestPrefix), current)
		newNode.addChild4(left2.key.At(int(depth)+longestPrefix), newLeftNode)
		replaceRef(curNode, newNode)
		return value, false
	}

	n := current.node()
	if n.partialLen > 0 {
		prefixMismatchedIdx := current.prefixMismatch(key, depth)
		if int(prefixMismatchedIdx) >= n.partialLen {
			depth += uint32(n.partialLen)
			goto RecurseSearch
		}

		// prefix lazy extend
		// newSharedArtNode for shared prefix
		newSharedArtNode := newNode4()
		newSharedArtNode.node().setPrefix(n.partial[:min(MaxPrefixLen, prefixMismatchedIdx)], int(prefixMismatchedIdx))
		if n.partialLen <= MaxPrefixLen {
			newSharedArtNode.addChild(n.partial[prefixMismatchedIdx], current)
			n.partialLen -= int(prefixMismatchedIdx + 1)
			if n.partialLen > 0 {
				copy(
					n.partial[:],
					n.partial[prefixMismatchedIdx+1:(int(prefixMismatchedIdx+1)+min(MaxPrefixLen, n.partialLen))])
			}
		} else {
			n.partialLen -= int(prefixMismatchedIdx + 1)
			l := leftmost(current)
			newSharedArtNode.addChild(l.key.At(int(depth+prefixMismatchedIdx)), current)
			if n.partialLen > 0 {
				copy(
					n.partial[:],
					l.key[depth+prefixMismatchedIdx:int(depth+prefixMismatchedIdx)+min(MaxPrefixLen, n.partialLen)],
				)
			}

		}
		newSharedArtNode.addChild(key.At(int(depth+prefixMismatchedIdx)), newLeaf(key, value))

		replaceRef(curNode, newSharedArtNode)
		return value, false
	}

RecurseSearch:
	// while prefixMismatchedIdx > node prefix len
	found := current.findChild(key.At(int(depth)))
	if *found != nil {
		return art.recursiveInsert(found, key, value, depth+1)
	}
	// just add a leaf node
	current.addChild(key.At(int(depth)), newLeaf(key, value))
	return value, false

}

func (art *artTree) recursiveRemove(curNode **artNode, key Key, depth uint32) (Value, bool) {
	current := *curNode
	if current == nil {
		return nil, false
	}

	// leaf node
	if current.isLeaf() {
		if current.leaf().Match(key) {
			replaceRef(curNode, nil)
			return current.leaf().value, true
		}
		return nil, false
	}

	n := current.node()
	// handling inner node
	if n.partialLen > 0 {
		prefixLen := current.checkPrefix(key, depth)
		if prefixLen != uint32(min(MaxPrefixLen, n.partialLen)) {
			return nil, false
		}
		depth += uint32(n.partialLen)
	}

	// find a child node
	child, idxOrChar := current.findChildAndIdx(key.At(int(depth)))
	if child == nil {
		return nil, false
	}

	if child.isLeaf() {
		if child.leaf().Match(key) {
			current.removeChildAt(byte(idxOrChar))
			return child.leaf().value, true
		}
		return nil, false
	}
	return art.recursiveRemove(&child, key, depth+1)
}
