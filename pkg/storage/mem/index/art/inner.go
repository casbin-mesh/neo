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
	"fmt"
)

func comparePrefix(k1, k2 []byte, depth int) int {
	idx, limit := depth, min(len(k1), len(k2))
	for ; idx < limit; idx++ {
		if k1[idx] != k2[idx] {
			break
		}
	}

	return idx - depth
}

func (n *inner[T]) Kind() Kind {
	return n.node.Kind()
}

func (n *inner[T]) leftmost() node[T] {
	return n.node.leftmost()
}

// prefixMismatch returns the index at which the prefix mismatched
func (n *inner[T]) prefixMismatch(key Key, depth int) (idx int) {
	maxCmp := min(min(maxPrefixLen, n.prefixLen), len(key)-depth)
	for ; idx < maxCmp; idx++ {
		if n.prefix[idx] != key[depth+idx] {
			return idx // mismatch
		}
	}

	// check the leftmost(minimum) node
	//        a (node4)        <----------- a.prefixMismatch("a leaf",0)
	//      x
	//     x
	//    x
	//   x
	// a leaf node here (leaf) <----------- compare the key of leftmost node
	if n.prefixLen > maxPrefixLen {
		l, ok := n.leftmost().(*leaf[T])
		if !ok {
			fmt.Printf("prefixMismatch got an incorrect leftmost node:%v\n", l)
			return idx
		}
		maxCmp = min(len(l.key), len(key)) - depth
		for ; idx < maxCmp; idx++ {
			if l.key[idx+depth] != key[depth+idx] {
				return idx
			}
		}
	}
	return
}

func (n *inner[T]) setPrefix(key []byte, len int) {
	n.prefixLen = len
	copy(n.prefix[:], key[:min(len, maxPrefixLen)])
}

func (n *inner[T]) insert(l *leaf[T], depth int, parent *olock, parentVersion uint64) (node[T], bool, bool) {
	for {
		version, obsolete := n.lock.RLock()
		if obsolete {
			return n, true, false
		}
		//prefixMismatchedIdx := comparePrefix(n.prefix[:n.prefixLen], l.key, 0, depth)
		prefixMismatchedIdx := n.prefixMismatch(l.key, depth)

		if prefixMismatchedIdx < n.prefixLen {
			if parent.Upgrade(parentVersion, nil) {
				return nil, true, false
			}
			if n.lock.Upgrade(version, parent) {
				return nil, true, false
			}

			// lazy expend
			//
			//     		this_is_a_long_prefix (current node)   <----- try to insert a leaf: this_is_leaf
			//  		*
			//  	*
			//  *
			// (1) index char
			// this_is_a_long_prefix1 (leaf)

			//     		this_is_ (node) 				  		   <----- new shared node
			//  		*           *
			//  	*		   	   		*
			// 	*							*
			// (l) index char				(a) index char
			// this_is_leaf (leaf)  		a_long_prefix (node)   <---- expanded node (current node) {prefix-sharedPrefix}
			// 								*
			// 							*
			// 						*
			// 					*
			// 				(1) index char
			//  			this_is_a_long_prefix1 (leaf)

			// current node will as child of n.node
			current := &inner[T]{
				node:      n.node,
				prefixLen: n.prefixLen,
			}
			// make a copy here
			copy(current.prefix[:], n.prefix[:])

			// n.node as a shared node
			n.node = &node4[T]{}
			// set prefix
			n.setPrefix(current.prefix[:min(maxPrefixLen, prefixMismatchedIdx)], prefixMismatchedIdx)

			if current.prefixLen <= maxPrefixLen {
				current.prefixLen -= prefixMismatchedIdx + 1
				n.node.addChild(current.prefix[prefixMismatchedIdx], current)
				// set current node's  prefix to {prefix - sharedPrefix}
				if current.prefixLen > 0 {
					copy(
						current.prefix[:],
						current.prefix[prefixMismatchedIdx+1:min(prefixMismatchedIdx+1+min(maxPrefixLen, current.prefixLen), maxPrefixLen)],
					)
				}
			} else { // prefixMismatchedId > maxPrefixLen
				current.prefixLen -= prefixMismatchedIdx + 1
				leftmost := current.leftmost().(*leaf[T])
				n.node.addChild(leftmost.key.At(depth+prefixMismatchedIdx), current)
				// set current node's prefix to {leftmost prefix - sharedPrefix}
				if current.prefixLen > 0 {
					copy(
						current.prefix[:],
						leftmost.key[depth+prefixMismatchedIdx+1:depth+prefixMismatchedIdx+1+min(maxPrefixLen, current.prefixLen)],
					)
				}
			}
			// add
			n.node.addChild(l.key.At(depth+prefixMismatchedIdx), l)

			n.lock.Unlock()
			parent.Unlock()
			return n, false, false
		}

		nextDepth := depth + n.prefixLen
		idx, next := n.node.child(l.key.At(nextDepth))

		if next == nil {
			if n.lock.Upgrade(version, nil) {
				continue
			}
			if parent.RUnlock(parentVersion, &n.lock) {
				return n, true, false
			}
			if n.node.full() {
				n.node = n.node.grow()
			}
			n.node.addChild(l.key.At(nextDepth), l)
			n.lock.Unlock()
			return n, false, false
		}
		if parent.RUnlock(parentVersion, nil) {
			return n, true, false
		}
		if next.isLeaf() {
			if n.lock.Upgrade(version, nil) {
				continue
			}

			replacement, _, updated := next.insert(l, nextDepth+1, &n.lock, version)
			n.node.replace(idx, replacement)
			n.lock.Unlock()
			return n, false, updated
		}

		_, restart, updated := next.insert(l, nextDepth+1, &n.lock, version)
		if restart {
			continue
		}
		return n, false, updated
	}
}

func (n *inner[T]) del(key Key, depth int, parent *olock, parentVersion uint64, parentUpdate func(node[T])) (deleted, restart bool, deletedNode node[T]) {
	for {
		version, obsolete := n.lock.RLock()
		if obsolete {
			return false, true, deletedNode
		}

		cmp := n.checkPrefix(key, depth)
		if cmp != min(n.prefixLen, maxPrefixLen) {
			// key is not found, check for concurrent writes and exit
			if n.lock.RUnlock(version, nil) {
				continue
			}
			return false, parent.RUnlock(parentVersion, nil), deletedNode
		}

		nextDepth := depth + n.prefixLen
		idx, next := n.node.child(key.At(nextDepth))
		if next == nil {
			// key is not found, check for concurrent writes and exit
			if n.lock.RUnlock(version, nil) {
				continue
			}
			return false, parent.RUnlock(parentVersion, nil), deletedNode
		}

		if l, isLeaf := next.(*leaf[T]); isLeaf && l.cmp(key) {
			_, isNode4 := n.node.(*node4[T])
			min := n.node.min()
			if isNode4 && min {
				// update parent pointer. current node will be collapsed.
				if parent.Upgrade(parentVersion, nil) {
					return false, true, deletedNode
				}
				if n.lock.Upgrade(version, parent) {
					// need to update parent version
					return false, true, deletedNode
				}
				deletedNode = n.node.replace(idx, nil)
				// get the left node
				leftB, left := n.node.next(nil)
				left.addPrefixBefore(n, leftB)
				parentUpdate(left)

				n.lock.Unlock()
				parent.Unlock()
				return true, false, deletedNode
			}
			// local change. parent lock won't be required
			if n.lock.Upgrade(version, nil) {
				continue
			}
			if parent.RUnlock(parentVersion, &n.lock) {
				return false, true, deletedNode
			}
			deletedNode = n.node.replace(idx, nil)
			if min && !isNode4 {
				n.node = n.node.shrink()
			}
			n.lock.Unlock()
			return true, false, deletedNode
		} else if isLeaf {
			// key is not found. check for concurrent writes and exit
			if n.lock.RUnlock(version, nil) {
				continue
			}
			return false, parent.RUnlock(parentVersion, nil), deletedNode
		}

		if parent.RUnlock(parentVersion, nil) {
			return false, true, deletedNode
		}

		if deleted, restart, deletedNode = next.del(key, nextDepth+1, &n.lock, version, func(rn node[T]) {
			n.node.replace(idx, rn)
		}); restart {
			continue
		}
		return deleted, false, deletedNode
	}
}

// checkPrefix Returns the number of prefix characters shared between
// the key and node.
func (n *inner[T]) checkPrefix(key Key, depth int) int {
	maxCmp := min(min(maxPrefixLen, n.prefixLen), len(key)-depth)
	idx := 0
	for ; idx < maxCmp; idx++ {
		if n.prefix[idx] != key[depth+idx] {
			return idx
		}
	}
	return idx
}

func (n *inner[T]) get(key Key, depth int, parent *olock, parentVersion uint64) (value T, found bool, restart bool) {
	for {
		version, obsolete := n.lock.RLock()
		if obsolete || parent.RUnlock(parentVersion, nil) {
			return value, false, true
		}
		prefixLen := n.checkPrefix(key, depth)
		if prefixLen != min(n.prefixLen, maxPrefixLen) {
			if n.lock.RUnlock(version, nil) {
				continue
			}
			return value, false, false
		}

		nextDepth := depth + n.prefixLen
		_, next := n.node.child(key.At(nextDepth))

		if next == nil {
			if n.lock.RUnlock(version, nil) {
				continue
			}
			return value, false, false
		}
		if next.isLeaf() {
			value, found, _ = next.get(key, nextDepth+1, &n.lock, version)
			if n.lock.RUnlock(version, nil) {
				continue
			}
			return value, found, false
		}
		value, found, restart = next.get(key, nextDepth+1, &n.lock, version)
		if restart {
			continue
		}
		return value, found, false
	}
}

func (n *inner[T]) walk(w walkFn[T], i int) bool {
	//TODO implement me
	panic("implement me")
}

func memcpy[T any](dst []T, src []T, len int) {
	copy(dst[:], src[:len])
}

func (n *inner[T]) addPrefixBefore(node *inner[T], key byte) {
	// new prefix: { node prefix } { key } { n(this) prefix }
	prefixCount := min(maxPrefixLen, node.prefixLen+1)
	memcpy(n.prefix[prefixCount:], n.prefix[:], min(n.prefixLen, maxPrefixLen-prefixCount))
	memcpy(n.prefix[:], node.prefix[:], min(prefixCount, node.prefixLen))
	if node.prefixLen < maxPrefixLen {
		n.prefix[prefixCount-1] = key
	}
	n.prefixLen += node.prefixLen + 1
}

//func (n *inner[T]) inherit(prefix [maxPrefixLen]byte, prefixLen int) node[T] {
//	// two cases for inheritance of the prefix
//	// 1. new prefixLen is <= max prefix len
//	total := n.prefixLen + prefixLen
//	if total <= maxPrefixLen {
//		copy(prefix[prefixLen:], n.prefix[:])
//		n.prefix = prefix
//		n.prefixLen = total
//		return n
//	}
//	// 2. >= max prefix len
//	// resplit prefix, first part should have 8-byte length
//	// second - leftover
//	// pointer should use 9th byte
//	// see long keys test
//	nn := &inner[T]{
//		node: &node4[T]{},
//	}
//	nn.prefix = prefix
//	nn.prefixLen = min(prefixLen, maxPrefixLen)
//	copy(nn.prefix[nn.prefixLen:], n.prefix[:])
//
//	n.prefixLen = total - maxPrefixLen - 1
//	kbyte := n.prefix[maxPrefixLen-min(maxPrefixLen, prefixLen)]
//	copy(n.prefix[:], n.prefix[maxPrefixLen-min(maxPrefixLen, prefixLen)+1:])
//	nn.node.addChild(kbyte, n)
//	return nn
//}

func (n *inner[T]) String() string {
	return fmt.Sprintf("inner[%x]%s", n.prefix[:n.prefixLen], n.node)
}

func (n *inner[T]) isLeaf() bool {
	return false
}
