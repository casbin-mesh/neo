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
	"github.com/casbin-mesh/neo/pkg/neo/index"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newIndex[T any]() index.Index[T] {
	return index.New[T](index.Options{})
}

func TestNewInMemMeta(t *testing.T) {
	txn := newIndex[any]().NewTransactionAt(1, true)
	meta := NewInMemMeta(txn)
	assert.NotNil(t, meta)
}

func metaHelper(t *testing.T) (index index.Index[any], ts uint64) {
	index = newIndex[any]()
	txn := index.NewTransactionAt(1, true)
	meta := NewInMemMeta(txn)
	assert.NotNil(t, meta)

	// insert first
	id, err := meta.NewDb("test_namespace")
	assert.Equal(t, uint64(1), id)
	assert.Nil(t, err)

	// duplicated namespace
	_, err = meta.NewDb("test_namespace")
	assert.Equal(t, ErrDbExists, err)

	// insert another namespace
	id, err = meta.NewDb("test_namespace_2")
	assert.Equal(t, uint64(2), id)
	assert.Nil(t, err)

	// commit it
	assert.Nil(t, meta.CommitAt(2))
	return index, 2
}

func TestInMemMeta_NewDb(t *testing.T) {
	txn := newIndex[any]().NewTransactionAt(1, true)
	meta := NewInMemMeta(txn)
	assert.NotNil(t, meta)

	// insert first
	id, err := meta.NewDb("test_namespace")
	assert.Equal(t, uint64(1), id)
	assert.Nil(t, err)

	// get
	id, err = meta.GetDBId("test_namespace")
	assert.Equal(t, uint64(1), id)
	assert.Nil(t, err)

	// duplicated namespace
	_, err = meta.NewDb("test_namespace")
	assert.Equal(t, ErrDbExists, err)

	// insert another namespace
	id, err = meta.NewDb("test_namespace_2")
	assert.Equal(t, uint64(2), id)
	assert.Nil(t, err)

	// get
	id, err = meta.GetDBId("test_namespace_2")
	assert.Equal(t, uint64(2), id)
	assert.Nil(t, err)

	// commit it
	assert.Nil(t, meta.CommitAt(2))
}

type GetSet struct {
	namespace string
	id        uint64
}

func TestInMemMeta_GetDBId(t *testing.T) {
	index, readTs := metaHelper(t)
	txn := index.NewTransactionAt(readTs, false)
	meta := NewInMemMeta(txn)
	sets := []GetSet{{namespace: "test_namespace", id: 1}, {namespace: "test_namespace_2", id: 2}}
	for _, set := range sets {
		id, err := meta.GetDBId(set.namespace)
		assert.Equal(t, set.id, id)
		assert.Nil(t, err)
	}
}
