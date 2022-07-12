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

type node256[T any] struct {
	lth      uint16
	children [256]node[T]
}

func (n *node256[T]) Kind() Kind {
	return Node256
}

func (n *node256[T]) leftmost() (v node[T]) {
	idx := 0
	for ; n.children[idx] == nil; idx++ {
	}
	return n.children[idx].leftmost()
}

func (n *node256[T]) child(k byte) (int, node[T]) {
	return int(k), n.children[k]
}

func (n *node256[T]) next(k *byte) (byte, node[T]) {
	for b, child := range n.children {
		if (k == nil || byte(b) > *k) && child != nil {
			return byte(b), child
		}
	}
	return 0, nil
}

func (n *node256[T]) prev(k *byte) (byte, node[T]) {
	for idx := n.lth - 1; idx >= 0; idx-- {
		b := byte(idx)
		child := n.children[idx]
		if (k == nil || b < *k) && child != nil {
			return b, child
		}
	}
	return 0, nil
}

func (n *node256[T]) replace(idx int, child node[T]) (old node[T]) {
	old = n.children[byte(idx)]
	n.children[byte(idx)] = child
	if child == nil {
		n.lth--
	}
	return
}

func (n *node256[T]) full() bool {
	return n.lth == 256
}

func (n *node256[T]) addChild(k byte, child node[T]) {
	n.children[k] = child
	n.lth++
}

func (n *node256[T]) grow() inode[T] {
	return nil
}

func (n *node256[T]) min() bool {
	return n.lth <= 49
}

func (n *node256[T]) shrink() inode[T] {
	nn := &node48[T]{
		lth: uint8(n.lth),
	}
	var index uint16
	for i := range n.children {
		if n.children[i] == nil {
			continue
		}
		index++
		nn.keys[i] = index
		nn.children[index-1] = n.children[i]
	}
	return nn
}

func (n *node256[T]) walk(fn walkFn[T], depth int) bool {
	for _, child := range n.children {
		if child != nil {
			if !child.walk(fn, depth) {
				return false
			}
		}
	}
	return true
}

func (n *node256[T]) String() string {
	var b bytes.Buffer
	_, _ = b.WriteString("n256[")
	encoder := hex.NewEncoder(&b)
	for i := range n.children {
		if n.children[i] != nil {
			_, _ = encoder.Write([]byte{byte(i)})
		}
	}
	_, _ = b.WriteString("]")
	return b.String()
}
