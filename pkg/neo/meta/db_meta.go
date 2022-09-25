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
	"fmt"
	"github.com/casbin-mesh/neo/pkg/db"
)

type dbMeta struct {
	db.Txn
}

func DBIsErrNotFound(err error) bool {
	return err == db.ErrKeyNotFound
}

func (i *dbMeta) newMeta(key []byte, idGen nextIdGen) (uint64, error) {
	var nextId uint64

	_, err := i.Get(key)
	if !DBIsErrNotFound(err) {
		return 0, fmt.Errorf("%s already exists", key)
	}

	if nextId, err = idGen(); err != nil {
		return 0, err
	}

	if err = i.Set(key, uint642Bytes(nextId)); err != nil {
		return 0, err
	}
	return nextId, nil
}

func (i *dbMeta) getMeta(key []byte) (uint64, error) {
	value, err := i.Get(key)
	if DBIsErrNotFound(err) {
		return 0, ErrKeyNotExists
	}
	buf, err := value.ValueCopy(nil)
	if err != nil {
		return 0, err
	}
	return byte2uint64(buf), nil
}

// incUint64 increases the value for key in index by step, returns increased value.
func (i *dbMeta) incUint64(key []byte, step uint64) (uint64, error) {
	old, err := i.getMeta(key)
	if DBIsErrNotFound(err) {
		err = i.Set(key, uint642Bytes(step))
		if err != nil {
			return 0, err
		}
		return step, nil
	}
	val := old + step
	if err = i.Set(key, uint642Bytes(val)); err != nil {
		return 0, err
	}
	return val, nil
}

func (i *dbMeta) CommitAt(commitTs uint64) error {
	return i.Txn.CommitAt(commitTs, nil)
}

func (i *dbMeta) Rollback() {
	i.Txn.Discard()
}

func NewDbMeta(txn db.Txn) ReaderWriter {
	return NewMeta(&dbMeta{txn})
}
