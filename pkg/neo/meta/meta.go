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

package meta

import (
	"encoding/binary"
	"errors"
	"github.com/casbin-mesh/neo/pkg/neo/codec"
)

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

type Base interface {
	newMeta(key []byte, idGen nextIdGen) (uint64, error)
	getMeta(key []byte) (uint64, error)
	incUint64(key []byte, step uint64) (uint64, error)
	CommitAt(commitTs uint64) error
	Rollback()
}

func uint642Bytes(v uint64) []byte {
	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], v)
	return data[:]
}

func byte2uint64(buf []byte) uint64 {
	return binary.LittleEndian.Uint64(buf)
}

type meta struct {
	Base
}

func (m meta) GetDBId(namespace string) (uint64, error) {
	return m.getMeta(codec.MetaKey(namespace))
}

func (m meta) GetTableId(did uint64, table string) (uint64, error) {
	return m.getMeta(codec.TableKey(did, table))
}

func (m meta) GetIndexId(tid uint64, index string) (uint64, error) {
	return m.getMeta(codec.IndexKey(tid, index))
}

func (m meta) GetMatcherId(did uint64, matcher string) (uint64, error) {
	return m.getMeta(codec.MatcherKey(did, matcher))
}

func (m meta) GetColumnId(tid uint64, column string) (uint64, error) {
	return m.getMeta(codec.ColumnKey(tid, column))
}

func (m meta) NewDb(namespace string) (uint64, error) {
	return m.newMeta(codec.MetaKey(namespace), func() (uint64, error) {
		return m.incUint64(mNextGlobalIDKey, 1)
	})
}

func (m meta) NewTable(did uint64, tableName string) (tableId uint64, err error) {
	return m.newMeta(codec.TableKey(did, tableName), func() (uint64, error) {
		return m.incUint64(mNextGlobalTableIDKey, 1)
	})
}

func (m meta) NewIndex(tid uint64, indexName string) (indexId uint64, err error) {
	return m.newMeta(codec.IndexKey(tid, indexName), func() (uint64, error) {
		return m.incUint64(mNextGlobalIndexIDKey, 1)
	})
}

func (m meta) NewMatcher(did uint64, matcher string) (matcherId uint64, err error) {
	return m.newMeta(codec.MatcherKey(did, matcher), func() (uint64, error) {
		return m.incUint64(mNextGlobalMatcherIDKey, 1)
	})
}

func (m meta) NewColumn(tid uint64, column string) (columnId uint64, err error) {
	return m.newMeta(codec.ColumnKey(tid, column), func() (uint64, error) {
		return m.incUint64(mNextGlobalColumnIDKey, 1)
	})
}

func (m meta) CommitAt(commitTs uint64) error {
	return m.Base.CommitAt(commitTs)
}

func (m meta) Rollback() {
	m.Base.Rollback()
}

func NewMeta(base Base) ReaderWriter {
	return &meta{base}
}
