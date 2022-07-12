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
	"encoding/hex"
)

type node48[T any] struct {
	lth      uint8
	keys     [256]uint16
	children [48]node[T]
}

func (n *node48[T]) Kind() Kind {
	return Node48
}

func (n *node48[T]) leftmost() (v node[T]) {
	idx := 0
	for ; n.keys[idx] == 0; idx++ {

	}
	return n.children[n.keys[idx]-1].leftmost()
}

func (n *node48[T]) child(k byte) (int, node[T]) {
	idx := n.keys[k]
	if idx == 0 {
		return 0, nil
	}
	return int(k), n.children[idx-1]
}

func (n *node48[T]) next(k *byte) (byte, node[T]) {
	for b, idx := range n.keys {
		if (k == nil || byte(b) > *k) && idx != 0 {
			return byte(b), n.children[idx-1]
		}
	}
	return 0, nil
}

func (n *node48[T]) prev(k *byte) (byte, node[T]) {
	for b := n.lth - 1; b >= 0; b-- {
		idx := n.keys[b]
		if (k == nil || byte(b) < *k) && idx != 0 {
			return byte(b), n.children[idx]
		}
	}
	return 0, nil
}

func (n *node48[T]) full() bool {
	return n.lth == 48
}

func (n *node48[T]) addChild(k byte, child node[T]) {
	for idx, existing := range n.children {
		if existing == nil {
			n.keys[k] = uint16(idx + 1)
			n.children[idx] = child
			n.lth++
			return
		}
	}
	panic("no empty slots")
}

func (n *node48[T]) grow() inode[T] {
	nn := &node256[T]{
		lth: uint16(n.lth),
	}
	for b, i := range n.keys {
		if i == 0 {
			continue
		}
		nn.children[b] = n.children[i-1]
	}
	return nn
}

func (n *node48[T]) replace(k int, child node[T]) (old node[T]) {
	idx := n.keys[k]
	if idx == 0 {
		panic("replace can't be called for idx=0")
	}
	old = n.children[idx-1]
	n.children[idx-1] = child
	if child == nil {
		n.keys[k] = 0
		n.lth--
	}
	return
}

func (n *node48[T]) min() bool {
	return n.lth <= 17
}

func (n *node48[T]) shrink() inode[T] {
	nn := &node16[T]{
		lth: n.lth,
	}
	nni := 0
	for i, idx := range n.keys {
		if idx == 0 {
			continue
		}
		child := n.children[idx-1]
		if child != nil {
			nn.keys[nni] = byte(i)
			nn.children[nni] = child
			nni++
		}
	}
	return nn
}

func (n *node48[T]) walk(fn walkFn[T], depth int) bool {
	for _, child := range n.children {
		if child != nil {
			if !child.walk(fn, depth) {
				return false
			}
		}
	}
	return true
}

func (n *node48[T]) String() string {
	var b bytes.Buffer
	_, _ = b.WriteString("n48[")
	encoder := hex.NewEncoder(&b)
	for i, idx := range n.keys {
		if idx == 0 {
			continue
		}
		child := n.children[idx-1]
		if child != nil {
			_, _ = encoder.Write([]byte{byte(i)})
		}
	}
	_, _ = b.WriteString("]")
	return b.String()
}
