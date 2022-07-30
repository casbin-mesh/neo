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

package db

import "github.com/casbin-mesh/neo/pkg/db/adapter"

type Item interface {
	// KeyCopy returns a copy of the key of the item, writing it to dst slice.
	// If nil is passed, or capacity of dst isn't sufficient, a new slice would be allocated and
	// returned.
	KeyCopy(dst []byte) []byte

	// ValueCopy returns a copy of the value of the item from the value log, writing it to dst slice.
	// If nil is passed, or capacity of dst isn't sufficient, a new slice would be allocated and
	// returned. Tip: It might make sense to reuse the returned slice as dst argument for the next call.
	//
	// This function is useful in long running iterate/update transactions to avoid a write deadlock.
	// See Github issue: https://github.com/dgraph-io/badger/issues/315
	ValueCopy(dst []byte) ([]byte, error)
}

type Iterator interface {
	Item() Item
	Valid() bool
	ValidForPrefix(b []byte) bool
	Close()
	Next()
	Seek(key []byte)
}

type Txn interface {
	// CommitAt commits the transaction, following the same logic as Commit(), but
	// at the given commit timestamp. This will panic if not used with managed transactions.
	//
	// This is only useful for databases built on top of Badger (like Dgraph), and
	// can be ignored by most users.
	CommitAt(commitTs uint64, callback func(error)) error
	Discard()
	Set([]byte, []byte) error
	Delete([]byte) error
	Get([]byte) (Item, error)
	NewKeyIterator(key []byte, iterOpt adapter.IteratorOptions) Iterator
	NewIterator(iterOpt adapter.IteratorOptions) Iterator
}

type DB interface {
	// NewTransactionAt follows the same logic as DB.NewTransaction(), but uses the
	// provided read timestamp.
	//
	// This is only useful for databases built on top of Badger (like Dgraph), and
	// can be ignored by most users.
	NewTransactionAt(readTs uint64, update bool) Txn
	// SetDiscardTs sets a timestamp at or below which, any invalid or deleted
	// versions can be discarded from the LSM tree, and thence from the value log to
	// reclaim disk space. Can only be used with managed transactions.
	SetDiscardTs(ts uint64)
	Close() error
}
