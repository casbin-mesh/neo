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

import "fmt"

type node16[T any] struct {
	lth      uint8
	keys     [16]byte
	children [16]node[T]
}

func (n *node16[T]) Kind() Kind {
	return Node16
}

func (n *node16[T]) leftmost() (v node[T]) {
	if n.children[0] != nil {
		return n.children[0].leftmost()
	}
	return
}

func (n *node16[T]) index(k byte) int {
	for i, b := range n.keys {
		if k <= b {
			return i
		}
	}
	return int(n.lth)
}

func index(key *byte, nkey *[16]byte) (int, bool) {
	for i := range nkey {
		if nkey[i] == *key {
			return i, true
		}
	}
	return 0, false
}

func (n *node16[T]) child(k byte) (int, node[T]) {
	idx, exist := index(&k, &n.keys)
	if !exist {
		return 0, nil
	}
	return idx, n.children[idx]
}

func (n *node16[T]) next(k *byte) (byte, node[T]) {
	if k == nil {
		return n.keys[0], n.children[0]
	}
	for i, b := range n.keys {
		if b > *k {
			return b, n.children[i]
		}
	}
	return 0, nil
}

func (n *node16[T]) prev(k *byte) (byte, node[T]) {
	if k == nil {
		idx := n.lth - 1
		return n.keys[idx], n.children[idx]
	}
	for i := n.lth; i >= 0; i-- {
		idx := i - 1
		if n.keys[idx] < *k {
			return n.keys[idx], n.children[idx]
		}
	}
	return 0, nil
}

func (n *node16[T]) replace(idx int, child node[T]) (old node[T]) {
	old = n.children[idx]
	if child == nil {
		copy(n.keys[idx:], n.keys[idx+1:])
		copy(n.children[idx:], n.children[idx+1:])
		n.keys[n.lth-1] = 0
		n.children[n.lth-1] = nil
		n.lth--
	} else {
		n.children[idx] = child
	}
	return
}

func (n *node16[T]) full() bool {
	return n.lth == 16
}

func (n *node16[T]) addChild(k byte, child node[T]) {
	idx := n.index(k)
	copy(n.children[idx+1:], n.children[idx:])
	copy(n.keys[idx+1:], n.keys[idx:])
	n.keys[idx] = k
	n.children[idx] = child
	n.lth++
}

func (n *node16[T]) grow() inode[T] {
	nn := &node48[T]{
		lth: n.lth,
	}
	copy(nn.children[:], n.children[:])
	for i, child := range n.children {
		if child == nil {
			continue
		}
		nn.keys[n.keys[i]] = uint16(i) + 1
	}
	return nn
}

func (n *node16[T]) min() bool {
	return n.lth <= 5
}

func (n *node16[T]) shrink() inode[T] {
	nn := node4[T]{}
	copy(nn.keys[:], n.keys[:])
	copy(nn.children[:], n.children[:])
	nn.lth = n.lth
	return &nn
}

func (n *node16[T]) walk(fn walkFn[T], depth int) bool {
	for i := range n.children {
		if uint8(i) < n.lth {
			if !n.children[i].walk(fn, depth) {
				return false
			}
		}
	}
	return true
}

func (n *node16[T]) String() string {
	return fmt.Sprintf("n16[%x]", n.keys[:n.lth])
}
