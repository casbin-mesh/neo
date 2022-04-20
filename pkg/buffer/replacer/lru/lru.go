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

package lru

import (
	"container/list"
	"errors"
	"github.com/casbin-mesh/neo/pkg/buffer/replacer"
	"sync"
)

var (
	ErrExceedMaxCap = errors.New("exceeded max capability")
)

// TODO: is the replacer possible to be lock-free? rewrite to an atomic version
type lruReplacer struct {
	mu        sync.RWMutex
	cap       uint64
	m         map[uint64]*list.Element
	leastUsed *list.List
}

func (l *lruReplacer) Pin(frameId uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.exist(frameId) {
		return nil
	}
	l.leastUsed.Remove(l.m[frameId])
	delete(l.m, frameId)
	return nil
}

func (l *lruReplacer) Unpin(frameId uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.exist(frameId) {
		return nil
	}
	if l.leastUsed.Len() == int(l.cap) {
		return ErrExceedMaxCap
	}
	l.m[frameId] = l.leastUsed.PushFront(frameId)
	return nil
}

func (l *lruReplacer) Victim(frameId *uint64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.leastUsed.Len() == 0 {
		return false
	}
	elem := l.leastUsed.Back()
	*frameId = elem.Value.(uint64)
	delete(l.m, *frameId)
	l.leastUsed.Remove(elem)
	return true
}

func (l *lruReplacer) Size() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return uint64(l.leastUsed.Len())
}

func (l *lruReplacer) exist(frameId uint64) bool {
	_, ok := l.m[frameId]
	return ok
}

func NewLRUReplacer(cap uint64) replacer.Replacer {
	return &lruReplacer{
		cap:       cap,
		m:         make(map[uint64]*list.Element),
		leastUsed: list.New(),
		mu:        sync.RWMutex{},
	}
}
