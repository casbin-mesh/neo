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

import "sync/atomic"

type Tree[T any] struct {
	lock olock
	root node[T]
	size int64
}

func (t *Tree[T]) Insert(key Key, value T) (updated bool) {
	for {
		version, restart := t.lock.RLock()
		l := &leaf[T]{key: key, value: value}
		root := t.root
		if root == nil { // empty tree, then insert a leaf node
			if t.lock.Upgrade(version, nil) {
				continue // restart
			}
			t.root = l
			t.lock.Unlock()
			atomic.AddInt64(&t.size, 1)
			return
		}
		if _, ok := root.(*leaf[T]); ok {
			if t.lock.Upgrade(version, nil) {
				continue // restart
			}
			t.root, _, updated = root.insert(l, 0, &t.lock, version)
			t.lock.Unlock()
			if !updated {
				atomic.AddInt64(&t.size, 1)
			}
			return
		}
		_, restart, updated = root.insert(l, 0, &t.lock, version)
		if restart {
			continue
		}
		if !updated {
			atomic.AddInt64(&t.size, 1)
		}
		return
	}
}

func (t *Tree[T]) Search(key Key) (value T, found bool) {
	restart := false
	for {
		version, _ := t.lock.RLock()
		root := t.root
		if root == nil {
			if t.lock.RUnlock(version, nil) {
				continue
			}
			return value, false
		}
		value, found, restart = root.get(key, 0, &t.lock, version)
		if restart {
			continue
		}
		return value, found
	}
}

func (t *Tree[T]) Remove(key Key) (deleted bool, value T) {
	restart := false
	var deletedNode node[T]
	for {
		version, _ := t.lock.RLock()
		root := t.root
		if root == nil {
			if t.lock.RUnlock(version, nil) {
				continue
			}
			return false, value
		}

		l, isLeaf := root.(*leaf[T])
		if isLeaf && l != nil && l.cmp(key) { // remove root leaf node
			if t.lock.Upgrade(version, nil) {
				continue
			}
			value = l.value
			t.root = nil
			t.lock.Unlock()
			return true, value
		} else if isLeaf { // mismatch
			if t.lock.RUnlock(version, nil) {
				continue
			}
			return false, value
		}

		if deleted, restart, deletedNode = root.del(key, 0, &t.lock, version, func(rn node[T]) {
			t.root = rn
		}); restart {
			continue
		}
		if deleted {
			value = deletedNode.(*leaf[T]).value
			atomic.AddInt64(&t.size, -1)
		}
		return deleted, value
	}
}

func (t *Tree[T]) Empty() (empty bool) {
	for {
		version, _ := t.lock.RLock()
		empty = t.root == nil
		if t.lock.RUnlock(version, nil) {
			continue // restart
		}
		return
	}
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
