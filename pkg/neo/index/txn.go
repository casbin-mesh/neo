// Copyright 2022 The casbin-mesh Authors. All Rights Reserved.
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

package index

import (
	"errors"
	"github.com/casbin-mesh/neo/pkg/storage/mem/index/art"
	"sync/atomic"
)

type Txn interface {
}

type txn[T any] struct {
	readTs        uint64
	update        bool
	root          *art.Tree[*VersionChainHead[T]]
	decRef        func()
	pendingWrites map[string]*Value[T]
	discarded     bool
}

func (m *txn[T]) Discord() {
	defer func() {
		if !m.discarded {
			m.discarded = true
			m.decRef()
		}
	}()

	for key, _ := range m.pendingWrites {
		value, exists := m.root.Search(art.Key(key))
		if exists {
			head := value
			head.mu.Lock()
			head.next = head.next.next
			head.next.txn = 0
			head.mu.Unlock()
			delete(m.pendingWrites, key) // release resources
		}
	}
}

func (m txn[T]) getVersion(key []byte, txnId uint64) (v *Value[T], err error) {
	value, exists := m.root.Search(key)
	if exists {
		head := value.next
		for head != nil {
			if txnId < head.beginTs || head.uncommitted {
				head = head.next
				continue
			}
			break
		}
		v = head
		if v == nil {
			return v, ErrKeyNotExists
		}
		vTxnId := atomic.LoadUint64(&v.txn) // if its write lock is not held by another active transaction
		if vTxnId != 0 && vTxnId != txnId {
			return v, ErrAnotherTxnHeldWLock
		}

		readTs := atomic.LoadUint64(&v.readTs)
		for txnId > readTs {
			if atomic.CompareAndSwapUint64(&v.readTs, readTs, txnId) {
				break
			}
			readTs = atomic.LoadUint64(&v.readTs)
		}
		return v, nil
	}
	return
}

func (m txn[T]) newVersion(key []byte, txnId uint64, value T) (*Value[T], error) {
	head, exists := m.root.Search(key)
	// TODO: change the insert to atomic
	// TODO: add SearchOrInsert
	if !exists {
		vi := &Value[T]{
			txn:         txnId, //w-lock held
			value:       value,
			uncommitted: true,
		}
		m.root.Insert(key, &VersionChainHead[T]{next: vi})
		return vi, nil
	}
	previous := head.next
	prevTxnId := atomic.LoadUint64(&previous.txn)
	prevReadTs := atomic.LoadUint64(&previous.readTs)
	allowed := txnId > prevReadTs && // txnId is larger previous version's readTs
		atomic.CompareAndSwapUint64(&previous.txn, 0, txnId) // // no active transaction holds previous version write lock

	if allowed {
		vi := &Value[T]{
			txn:         txnId, //w-lock held
			next:        previous,
			uncommitted: true,
		}
		verHead := head
		verHead.mu.Lock()
		verHead.next = vi
		verHead.mu.Unlock()
		return vi, nil
	}

	if prevTxnId != 0 || prevReadTs > txnId {
		return nil, ErrFailedToAcquireWLock
	}
	return nil, ErrWriteConflicts
}

func (m txn[T]) Get(key []byte) (ret T, err error) {

	// read pending writes
	v, ok := m.pendingWrites[string(key)]
	if ok {
		return v.value, nil
	}

	// T is allowed to read version Ax
	// if its write lock is not held by another active transaction
	vi, err := m.getVersion(key, m.readTs)
	if err != nil {
		return ret, err
	}
	return vi.value, nil
}

var (
	ErrAnotherTxnHeldWLock  = errors.New("another txn held write-lock")
	ErrFailedToAcquireWLock = errors.New("failed to acquire write-lock")
	ErrWriteConflicts       = errors.New("write conflicts")
	ErrKeyNotExists         = errors.New("key not exists")
)

func (m txn[T]) Set(key []byte, value T) error {
	// update key
	// With MVTO, a transaction always updates the latest version of a tuple.
	// Transaction T creates a new version Bx+1 if
	//		(1) no active transaction holds Bx’s write lock and
	// 		(2) Tid is larger than Bx’s read-ts field.
	v, ok := m.pendingWrites[string(key)]
	if ok {
		v.value = value
		return nil
	}

	vi, err := m.newVersion(key, m.readTs, value)
	if err != nil {
		return err
	}
	m.pendingWrites[string(key)] = vi
	return nil
}

func (m *txn[T]) CommitAt(commitTs uint64, callback func(error)) error {
	defer func() {
		if !m.discarded {
			m.discarded = true
			m.decRef()
		}
	}()

	// When T commits,
	// the DBMS sets Bx+1’s begin-ts and end-ts fields to Tid and INF (respectively),
	// and Bx’s end-ts field to Tid.
	for key, wr := range m.pendingWrites {
		if wr.next != nil {
			wr.txn = 0
			wr.next.txn = 0
			wr.next.endTs = commitTs
		}
		wr.txn = 0
		wr.beginTs = commitTs
		wr.endTs = ^uint64(0)
		wr.uncommitted = false
		delete(m.pendingWrites, key)
	}
	return nil
}

func (m txn[T]) ReadTS() uint64 {
	return m.readTs
}
