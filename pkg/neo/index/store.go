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
	"github.com/casbin-mesh/neo/pkg/storage/mem/index/art"
	"sync/atomic"
)

type backend[T any] struct {
	// root storage backend,
	//
	// each leaf point to version chain header.
	// NOTES: the tree is NOT thread-safe
	// TODO(weny): change it to sync version and supports the generic
	root *art.Tree[*VersionChainHead[T]]

	// TODO: move it to epochs
	txnCnt uint64
}

type Options struct {
}

func New[T any](opts Options) *backend[T] {
	return &backend[T]{
		root: &art.Tree[*VersionChainHead[T]]{},
	}
}

func (s *backend[T]) incRef() {
	atomic.AddUint64(&s.txnCnt, uint64(1))
}

func (s *backend[T]) decRef() {
	atomic.AddUint64(&s.txnCnt, ^uint64(1))
}

func (s *backend[T]) NewTransactionAt(readTs uint64, update bool) *txn[T] {
	s.incRef()
	return &txn[T]{
		pendingWrites: make(map[string]*Value[T]),
		decRef:        s.decRef,
		root:          s.root,
		readTs:        readTs,
		update:        update,
	}
}
