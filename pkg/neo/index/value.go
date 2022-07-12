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

import "sync"

type VersionChainHead[T any] struct {
	next *Value[T]
	mu   sync.Mutex
}

// Value value and Timestamp Ordering header (MVTO)
type Value[T any] struct {
	// header
	// if txn is not zero, means the write-lock hold by the txn
	txn         uint64
	readTs      uint64
	beginTs     uint64
	endTs       uint64 // TODO: uses uint64.MAX to represent the +INF ?
	uncommitted bool

	// pointer to older version
	next *Value[T]
	// value
	value T
}
