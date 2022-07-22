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
	"fmt"
)

type LEAF[T any] struct {
	lf leaf[T]
}

type leaf[T any] struct {
	key   Key
	value T
}

func (l leaf[T]) Kind() Kind {
	return Leaf
}

func (l *leaf[T]) leftmost() node[T] {
	return l
}

func (l *leaf[T]) insert(other *leaf[T], depth int, parent *olock, parentVersion uint64) (value node[T], restart bool, updated bool) {
	if other.cmp(l.key) { // replace
		return other, false, true
	}

	longestPrefix := comparePrefix(l.key, other.key, depth)
	nn := &inner[T]{
		prefixLen: longestPrefix,
		node:      &node4[T]{},
	}
	nn.setPrefix(other.key[depth:], longestPrefix)

	nn.node.addChild(l.key.At(depth+longestPrefix), l)
	nn.node.addChild(other.key.At(depth+longestPrefix), other)

	return nn, false, false
}

func (l leaf[T]) del(bytes Key, i int, o *olock, u uint64, f func(node[T])) (bool, bool, node[T]) {
	panic("not needed")
}

func (l leaf[T]) get(key Key, i int, o *olock, u uint64) (value T, found bool, restart bool) {
	if l.cmp(key) {
		return l.value, true, false
	}
	return value, false, false
}

func (l *leaf[T]) walk(fn walkFn[T], depth int) bool {
	panic("implement me")
}

func (n *leaf[T]) addPrefixBefore(node *inner[T], key byte) {

}

//
//func (l *leaf[T]) inherit(prefix [maxPrefixLen]byte, prefixLen int) node[T] {
//	return l
//}

func (l leaf[T]) isLeaf() bool {
	return true
}

func (l leaf[T]) String() string {
	return fmt.Sprintf("leaf[%x]", l.key)
}

func (l *leaf[T]) cmp(other []byte) bool {
	return bytes.Compare(l.key, other) == 0
}
