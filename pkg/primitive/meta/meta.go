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
	"errors"
	"github.com/casbin-mesh/neo/pkg/db"
	"github.com/casbin-mesh/neo/pkg/primitive/utils"
)

type KeyValue interface {
	Key() []byte
	Value
}
type Value interface {
	ValueCopy([]byte) ([]byte, error)
}

type Reader interface {
	Get(namespace, name []byte) (Value, error)
	Discard() error
}

type Writer interface {
	Set(namespace []byte, schema KeyValue) error
	Delete(namespace, name []byte) error
	Commit() error
	Discard() error
}

type ReaderWriter interface {
	Reader
	Writer
}

func NewWriter(txn db.Txn) ReaderWriter {
	return &meta{txn: txn}
}

func NewReader(txn db.Txn) Reader {
	return &meta{txn: txn}
}

type meta struct {
	txn db.Txn
}

func (m meta) Delete(namespace, name []byte) error {
	key := utils.CString(namespace, name)
	return m.txn.Delete(key)
}

func (m meta) Get(namespace []byte, name []byte) (Value, error) {
	key := utils.CString(namespace, name)
	if item, err := m.txn.Get(key); err == nil {
		return item, nil
	}
	return nil, ErrNotExists
}

func (m meta) Set(namespace []byte, schema KeyValue) error {
	key := utils.CString(namespace, schema.Key())
	value, err := schema.ValueCopy(nil)
	if err != nil {
		return err
	}
	if err = m.txn.Set(key, value); err != nil {
		return err
	}
	return nil
}

func (m meta) Commit() error {
	return m.Commit()
}

func (m meta) Discard() error {
	return m.Discard()
}

var (
	ErrNotExists = errors.New("not exists")
)
