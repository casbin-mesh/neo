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

package badgerAdapter

import (
	"github.com/casbin-mesh/neo/pkg/db"
	adapter2 "github.com/casbin-mesh/neo/pkg/db/adapter"
	"github.com/dgraph-io/badger/v3"
)

type adapter struct {
	db *badger.DB
}

type txn struct {
	txn    *badger.Txn
	readTs uint64
}

type iter struct {
	*badger.Iterator
}

func (i *iter) Item() db.Item {
	return i.Iterator.Item()
}

func (t txn) CommitAt(commitTs uint64, callback func(error)) error {
	return t.txn.CommitAt(commitTs, callback)
}

func (t txn) Discard() {
	t.txn.Discard()
}

func (t txn) Commit() error {
	return t.txn.Commit()
}

func (t txn) ReadTs() uint64 {
	return t.readTs
}

func (t txn) CommitWith(cb func(err error)) {
	t.txn.CommitWith(cb)
}

func (t txn) Set(k []byte, v []byte) error {
	return t.txn.Set(k, v)
}

func (t txn) Delete(k []byte) (err error) {
	err = t.txn.Delete(k)
	if err == badger.ErrKeyNotFound {
		return db.ErrKeyNotFound
	}
	return err
}

func (t txn) NewKeyIterator(key []byte, iterOpt adapter2.IteratorOptions) db.Iterator {
	i := t.txn.NewKeyIterator(key, badger.IteratorOptions{
		PrefetchSize:   iterOpt.PrefetchSize,
		PrefetchValues: iterOpt.PrefetchValues,
		Reverse:        iterOpt.Reverse,
		AllVersions:    iterOpt.AllVersions,
		InternalAccess: iterOpt.InternalAccess,
		Prefix:         iterOpt.Prefix,
		SinceTs:        iterOpt.SinceTs,
	})
	return &iter{Iterator: i}
}

func (t txn) NewIterator(iterOpt adapter2.IteratorOptions) db.Iterator {
	i := t.txn.NewIterator(badger.IteratorOptions{
		PrefetchSize:   iterOpt.PrefetchSize,
		PrefetchValues: iterOpt.PrefetchValues,
		Reverse:        iterOpt.Reverse,
		AllVersions:    iterOpt.AllVersions,
		InternalAccess: iterOpt.InternalAccess,
		Prefix:         iterOpt.Prefix,
		SinceTs:        iterOpt.SinceTs,
	})
	return &iter{Iterator: i}
}

func (t txn) Get(k []byte) (db.Item, error) {
	if item, err := t.txn.Get(k); err != nil {
		if err == badger.ErrKeyNotFound {
			return item, db.ErrKeyNotFound
		}
		return item, err
	} else {
		return item, nil
	}
}

func (b adapter) NewTransactionAt(readTs uint64, update bool) db.Txn {
	t := b.db.NewTransactionAt(readTs, update)
	return &txn{txn: t, readTs: readTs}
}

func (b adapter) NewTransaction(update bool) db.Txn {
	t := b.db.NewTransaction(update)
	return &txn{txn: t}
}

func (b adapter) SetDiscardTs(ts uint64) {
	b.db.SetDiscardTs(ts)
}

func OpenManaged(opt badger.Options) (db.DB, error) {
	db, err := badger.OpenManaged(opt)
	if err != nil {
		return nil, err
	}
	return &adapter{db: db}, nil
}

func (b adapter) MaxVersion() uint64 {
	return b.db.MaxVersion()
}

func Open(opt badger.Options) (db.DB, error) {
	db, err := badger.Open(opt)
	if err != nil {
		return nil, err
	}
	return &adapter{db: db}, nil
}

func (b adapter) Close() error {
	return b.db.Close()
}
