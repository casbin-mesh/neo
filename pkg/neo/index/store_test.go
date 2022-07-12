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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	s := New[int](Options{})
	assert.NotNil(t, s)
}

func setHelper[T any](t *testing.T, txn *txn[T], key string, value T) {
	assert.Nil(t, txn.Set([]byte(key), value))
}

func TestTxn(t *testing.T) {
	t.Run("should not see uncommitted versions", func(t *testing.T) {
		s := New[int](Options{})
		txn1 := s.NewTransactionAt(1, true)
		setHelper[int](t, txn1, "hello", 1)
		setHelper[int](t, txn1, "alice", 2)
		setHelper[int](t, txn1, "bob", 3)

		txn2 := s.NewTransactionAt(2, false)
		_, err := txn2.Get([]byte("hello"))
		assert.Equal(t, ErrKeyNotExists, err)
		_, err = txn2.Get([]byte("alice"))
		assert.Equal(t, ErrKeyNotExists, err)
		_, err = txn2.Get([]byte("bob"))
		assert.Equal(t, ErrKeyNotExists, err)

	})
	t.Run("should see committed versions", func(t *testing.T) {
		s := New[int](Options{})
		txn1 := s.NewTransactionAt(1, true)
		setHelper[int](t, txn1, "hello", 1)
		setHelper[int](t, txn1, "alice", 2)
		setHelper[int](t, txn1, "bob", 3)
		assert.Nil(t, txn1.CommitAt(1, nil))

		// we should see the committed versions
		txn2 := s.NewTransactionAt(2, true)
		_, err := txn2.Get([]byte("hello"))
		assert.Nil(t, err)
		_, err = txn2.Get([]byte("alice"))
		assert.Nil(t, err)
		_, err = txn2.Get([]byte("bob"))
		assert.Nil(t, err)
	})
	t.Run("should failed to acquire w-lock", func(t *testing.T) {
		s := New[int](Options{})

		txn1 := s.NewTransactionAt(1, true)
		setHelper[int](t, txn1, "hello", 1)

		// due to txn1 still holding the lock
		txn2 := s.NewTransactionAt(2, true)
		err := txn2.Set([]byte("hello"), 2)
		assert.Equal(t, ErrFailedToAcquireWLock, err)
		err = txn2.Set([]byte("alice"), 2)
		assert.Nil(t, err)
	})
	t.Run("should failed to acquire w-lock", func(t *testing.T) {
		s := New[int](Options{})

		txn1 := s.NewTransactionAt(1, true)
		setHelper[int](t, txn1, "hello", 1)
		assert.Nil(t, txn1.CommitAt(1, nil))

		txn2 := s.NewTransactionAt(3, true)
		v, err := txn2.Get([]byte("hello"))
		assert.Nil(t, err)
		assert.Equal(t, 1, v)

		txn3 := s.NewTransactionAt(2, true)
		err = txn3.Set([]byte("hello"), 2)
		assert.Equal(t, ErrFailedToAcquireWLock, err)
	})

}
