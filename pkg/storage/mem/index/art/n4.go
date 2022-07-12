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

type node4[T any] struct {
	lth      uint8
	keys     [4]byte
	children [4]node[T]
}

func (n *node4[T]) Kind() Kind {
	return Node4
}

func (n *node4[T]) index(k byte) int {
	for i, b := range n.keys {
		if k <= b {
			return i
		}
	}
	return int(n.lth)
}

func (n *node4[T]) next(k *byte) (byte, node[T]) {
	if k == nil {
		return n.keys[0], n.children[0]
	}
	for idx, b := range n.keys {
		if b > *k {
			return b, n.children[idx]
		}
	}
	return 0, nil
}

func (n *node4[T]) prev(k *byte) (byte, node[T]) {
	if n.lth == 0 {
		return 0, nil
	}
	if k == nil {
		idx := n.lth - 1
		return n.keys[idx], n.children[idx]
	}
	for i := n.lth; i > 0; i-- {
		idx := i - 1
		if n.keys[idx] < *k {
			return n.keys[idx], n.children[idx]
		}
	}
	return 0, nil
}

func (n *node4[T]) leftmost() (v node[T]) {
	if n.children[0] != nil {
		return n.children[0].leftmost()
	}
	return
}

func (n *node4[T]) child(k byte) (int, node[T]) {
	idx := n.index(k)
	if uint8(idx) == n.lth {
		return 0, nil
	}
	if n.keys[idx] != k {
		return idx, nil
	}
	return idx, n.children[idx]
}

func (n *node4[T]) addChild(k byte, child node[T]) {
	idx := n.index(k)
	copy(n.children[idx+1:], n.children[idx:])
	copy(n.keys[idx+1:], n.keys[idx:])
	n.keys[idx] = k
	n.children[idx] = child
	n.lth++
}

func (n *node4[T]) replace(idx int, child node[T]) (old node[T]) {
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

func (n *node4[T]) full() bool {
	return n.lth == 4
}

func (n *node4[T]) grow() inode[T] {
	nn := &node16[T]{}
	nn.lth = n.lth
	copy(nn.keys[:], n.keys[:])
	copy(nn.children[:], n.children[:])
	return nn
}

func (n *node4[T]) min() bool {
	return n.lth <= 2
}

func (n *node4[T]) shrink() inode[T] {
	panic("can't shrink node4")
}

func (n *node4[T]) walk(fn walkFn[T], depth int) bool {
	for i := range n.children {
		if uint8(i) < n.lth {
			if !n.children[i].walk(fn, depth) {
				return false
			}
		}
	}
	return true
}

func (n *node4[T]) String() string {
	return fmt.Sprintf("n4[%x]", n.keys[:n.lth])
}
