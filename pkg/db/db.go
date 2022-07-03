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
	ValueCopy(dst []byte) []byte
}

type Txn interface {
	Discard()
	Commit() error
	CommitWith(cb func(err error))
	Set([]byte, []byte) error
	Delete([]byte) error
	Get([]byte) (Item, error)
}

type DB interface {
	NewTransaction(update bool) Txn
}