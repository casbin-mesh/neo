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
	"unsafe"
)

type artLeaf struct {
	value Value
	key   Key
}

func newLeaf(key Key, value Value) *artNode {
	return &artNode{
		kind: Leaf,
		ref:  unsafe.Pointer(&artLeaf{key: key.Clone(), value: value}),
	}
}

func (l *artLeaf) Match(key Key) bool {
	if key == nil || len(key) != len(l.key) {
		return false
	}
	return bytes.Compare(l.key, key) == 0
}

func (l *artLeaf) PartialMatch(key Key, depth uint32) bool {
	l1key, l2key := l.key, key
	idx, limit := depth, min(len(l1key), len(l2key))
	for ; idx < uint32(limit); idx++ {
		if l1key[idx] != l2key[idx] {
			break
		}
	}
	return int(idx-depth) > 0
}

func longestCommonPrefix(l1 *artLeaf, l2 *artLeaf, depth uint32) int {
	l1key, l2key := l1.key, l2.key
	idx, limit := depth, min(len(l1key), len(l2key))
	for ; idx < uint32(limit); idx++ {
		if l1key[idx] != l2key[idx] {
			break
		}
	}

	return int(idx - depth)
}
