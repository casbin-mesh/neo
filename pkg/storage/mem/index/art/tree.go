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

// checkPrefix Returns the number of prefix characters shared between the key and node.
func (art *artTree) checkPrefix(n *artNode, key []byte, depth int) (idx int) {
	maxCmp := min(min(int(n.partialLen), MaxPrefixLen), len(key)-depth)
	for ; idx < maxCmp; idx++ {
		if n.partial[idx] != key[depth+idx] {
			return
		}
	}
	return
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
		return leftmost(node.children[node.keys[idx]])

	case Node256:
		node := an.node256()
		idx := 255
		for ; node.children[idx] == nil; idx-- {
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
		for ; node.children[node.keys[idx]] == nil; idx-- {
		}
		return rightmost(node.children[node.keys[idx]])

	case Node256:
		node := an.node256()
		idx := 0
		for ; node.children[idx] == nil; idx++ {
		}
		return rightmost(node.children[idx])
	}

	return nil // that should never happen in normal case
}

func (art *artTree) Insert(key Key, value Value) (Value, bool) {
	old, updated := art.recursiveInsert(art.root, key, value, 0)
	if !updated {
		art.size++
	}
	return old, updated
}

func (art *artTree) Remove(key Key, value Value) (Value, bool) {
	value, deleted := art.recursiveRemove(art.root, key, 0)
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

		if current.partialLen > 0 {
			prefixLen := current.checkPrefix(key, depth)
			if prefixLen != uint32(min(current.partialLen, MaxPrefixLen)) {
				return nil, false
			}
			depth += uint32(current.partialLen)
		}

		next := current.findChild(key[depth])
		if next != nil {
			current = next
		} else {
			current = nil
		}
		depth++

	}
	return nil, false
}

func (art *artTree) recursiveInsert(curNode *artNode, key Key, value Value, depth uint32) (Value, bool) {
	current := curNode
	// if current is nil, insert a left node
	if current == nil {
		replaceRef(curNode, newLeaf(key, value))
		return nil, false
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
		newNode.setPrefix(key, longestPrefix)
		// it's safe to call add child directly
		newNode.addChild4(key[int(depth)+longestPrefix], current)
		newNode.addChild4(left2.key[int(depth)+longestPrefix], newLeftNode)
		replaceRef(curNode, newNode)
	}

	if current.partialLen > 0 {
		prefixMismatchedIdx := current.prefixMismatch(key, depth)
		if int(prefixMismatchedIdx) > current.partialLen {
			depth += uint32(current.partialLen)
			goto RecurseSearch
		}

		// prefix lazy extend
		// newSharedNode for shared prefix
		newSharedNode := newNode4()
		newSharedNode.setPrefix(current.partial[:min(MaxPrefixLen, prefixMismatchedIdx)], int(prefixMismatchedIdx))
		if current.partialLen <= MaxPrefixLen {
			newSharedNode.addChild(current.partial[prefixMismatchedIdx], current)
			current.partialLen -= int(prefixMismatchedIdx) + 1
			copy(current.partial[:], current.partial[prefixMismatchedIdx:min(MaxPrefixLen, current.partialLen)])
		} else {
			current.partialLen -= int(prefixMismatchedIdx) + 1
			l := leftmost(current)
			newSharedNode.addChild(l.key[depth+prefixMismatchedIdx], current)
			copy(current.partial[:], current.partial[prefixMismatchedIdx:min(MaxPrefixLen, current.partialLen)])
		}
		newSharedNode.addChild(key[depth+prefixMismatchedIdx], newLeaf(key, value))

		replaceRef(curNode, newSharedNode)
		return nil, false
	}

RecurseSearch:
	// while prefixMismatchedIdx > node prefix len
	found := current.findChild(key[depth])
	if found != nil {
		return art.recursiveInsert(found, key, value, depth+1)
	}
	// just add a leaf node
	current.addChild(key[depth], newLeaf(key, value))
	return nil, false

}

func (art *artTree) recursiveRemove(curNode *artNode, key Key, depth uint32) (Value, bool) {
	if curNode == nil {
		return nil, false
	}

	// leaf node
	if curNode.isLeaf() {
		if curNode.leafMatch(key) {
			replaceRef(curNode, nil)
			return curNode.leaf().value, true
		}
		return nil, false
	}

	// handling inner node
	if curNode.partialLen > 0 {
		prefixLen := curNode.checkPrefix(key, depth)
		if prefixLen != uint32(min(MaxPrefixLen, curNode.partialLen)) {
			return nil, false
		}
		depth += uint32(curNode.partialLen)
	}

	// find a child node
	child, idxOrChar := curNode.findChildAndIdx(key[depth])
	if child == nil {
		return nil, false
	}

	if child.isLeaf() {
		if child.leafMatch(key) {
			curNode.removeChildAt(byte(idxOrChar))
			return child.leaf().value, true
		}
		return nil, false
	}
	return art.recursiveRemove(child, key, depth+1)
}
