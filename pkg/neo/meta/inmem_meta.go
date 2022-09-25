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
	"fmt"
	"github.com/casbin-mesh/neo/pkg/neo/index"
)

func IsErrNotFound(err error) bool {
	return err == index.ErrKeyNotExists
}

type inMemMeta struct {
	index.Txn[any]
}

func (i *inMemMeta) CommitAt(commitTs uint64) error {
	return i.Txn.CommitAt(commitTs, nil)
}

func (i *inMemMeta) Rollback() {
	//TODO: TBD
	i.Discard()
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

func NewInMemMeta(index index.Txn[any]) ReaderWriter {
	return NewMeta(&inMemMeta{index})
}
