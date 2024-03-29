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

package meta

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/casbin-mesh/neo/pkg/neo/codec"
	"github.com/casbin-mesh/neo/pkg/neo/index"
)

type Reader interface {
	GetDBId(namespace string) (uint64, error)
	GetTableId(did uint64, table string) (uint64, error)
	GetIndexId(tid uint64, index string) (uint64, error)
	GetMatcherId(did uint64, matcher string) (uint64, error)
	GetColumnId(tid uint64, column string) (uint64, error)
}

type ReaderWriter interface {
	Reader
	NewDb(namespace string) (uint64, error)
	NewTable(did uint64, tableName string) (tableId uint64, err error)
	NewIndex(tid uint64, indexName string) (indexId uint64, err error)
	NewMatcher(did uint64, matcher string) (matcherId uint64, err error)
	NewColumn(tid uint64, column string) (columnId uint64, err error)

	CommitAt(commitTs uint64) error
	Rollback()
}

var (
	mNextGlobalIDKey        = []byte("g_NextGlobalID_db")
	mNextGlobalTableIDKey   = []byte("g_NextGlobalID_table")
	mNextGlobalIndexIDKey   = []byte("g_NextGlobalID_index")
	mNextGlobalMatcherIDKey = []byte("g_NextGlobalID_matcher")
	mNextGlobalColumnIDKey  = []byte("g_NextGlobalID_column")

	ErrKeyNotExists = errors.New("key does not exists")
	ErrUnknownType  = errors.New("unknown type")
	ErrDbExists     = errors.New("db already exists")
)

func IsErrNotFound(err error) bool {
	return err == index.ErrKeyNotExists
}

type inMemMeta struct {
	index.Txn[any]
}

func uint642Bytes(v uint64) [8]byte {
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], v)
	return data
}

func (i *inMemMeta) CommitAt(commitTs uint64) error {
	return i.Txn.CommitAt(commitTs, nil)
}

func (i *inMemMeta) Rollback() {
	//TODO: TBD
}

// incUint64 increases the value for key in index by step, returns increased value.
func (i *inMemMeta) incUint64(key []byte, step uint64) (uint64, error) {
	old, err := i.Get(key)
	if IsErrNotFound(err) {
		err = i.Set(key, step)
		if err != nil {
			return 0, err
		}
		return step, nil
	}
	if val, ok := old.(uint64); ok {
		val += step
		if err = i.Set(key, val); err != nil {
			return 0, err
		}
		return val, nil
	}
	return 0, ErrUnknownType
}

func (i *inMemMeta) NextGlobalId() (uint64, error) {
	return i.incUint64(mNextGlobalIDKey, 1)
}

func (i *inMemMeta) getMeta(key []byte) (uint64, error) {
	value, err := i.Get(key)
	if IsErrNotFound(err) {
		return 0, ErrKeyNotExists
	}
	if id, ok := value.(uint64); ok {
		return id, nil
	}
	return 0, ErrUnknownType
}

func (i *inMemMeta) GetDBId(namespace string) (uint64, error) {
	return i.getMeta(codec.MetaKey(namespace))
}

func (i *inMemMeta) GetTableId(did uint64, table string) (uint64, error) {
	return i.getMeta(codec.TableKey(did, table))
}

func (i *inMemMeta) GetIndexId(tid uint64, index string) (uint64, error) {
	return i.getMeta(codec.IndexKey(tid, index))
}

func (i *inMemMeta) GetMatcherId(did uint64, matcher string) (uint64, error) {
	return i.getMeta(codec.MatcherKey(did, matcher))
}

func (i *inMemMeta) GetColumnId(tid uint64, column string) (uint64, error) {
	return i.getMeta(codec.ColumnKey(tid, column))
}

type nextIdGen func() (uint64, error)

func (i *inMemMeta) newMeta(key []byte, idGen nextIdGen) (uint64, error) {
	var nextId uint64

	_, err := i.Get(key)
	if !IsErrNotFound(err) {
		return 0, fmt.Errorf("%s already exists", key)
	}

	if nextId, err = idGen(); err != nil {
		return 0, err
	}

	if err = i.Set(key, nextId); err != nil {
		return 0, err
	}
	return nextId, nil
}

func (i *inMemMeta) NewColumn(tid uint64, column string) (columnId uint64, err error) {
	return i.newMeta(codec.ColumnKey(tid, column), func() (uint64, error) {
		return i.incUint64(mNextGlobalColumnIDKey, 1)
	})
}

func (i *inMemMeta) NewDb(namespace string) (uint64, error) {
	return i.newMeta(codec.MetaKey(namespace), i.NextGlobalId)
}

func (i *inMemMeta) NewTable(did uint64, tableName string) (tableId uint64, err error) {
	return i.newMeta(codec.TableKey(did, tableName), func() (uint64, error) {
		return i.incUint64(mNextGlobalTableIDKey, 1)
	})
}

func (i *inMemMeta) NewIndex(tid uint64, indexName string) (indexId uint64, err error) {
	return i.newMeta(codec.IndexKey(tid, indexName), func() (uint64, error) {
		return i.incUint64(mNextGlobalIndexIDKey, 1)
	})
}

func (i *inMemMeta) NewMatcher(did uint64, matcher string) (matcherId uint64, err error) {
	return i.newMeta(codec.MatcherKey(did, matcher), func() (uint64, error) {
		return i.incUint64(mNextGlobalMatcherIDKey, 1)
	})
}

func NewInMemMeta(index index.Txn[any]) ReaderWriter {
	return &inMemMeta{index}
}
