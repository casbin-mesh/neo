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

package neo

import (
	"github.com/casbin-mesh/neo/pkg/db"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
	"github.com/casbin-mesh/neo/pkg/primitive/meta"
)

type mutation struct {
	ctx *mutationCtx
	txn db.Txn
}

type mutationCtx struct {
	namespace []byte
	metaRW    meta.ReaderWriter
}

type Value interface {
	ValueCopy() []byte
}

// ParseModel parses model string from user inputs
func (m *mutation) ParseModel(modelStr string) error {
	// 1. parses request schema definitions
	// 1.1 m.ctx.metaRW.Set(m.ctx.namespace,BSchema)

	// 2. parses policies schema definitions
	// 3. parses group schema definitions
	// 4. parses effect definitions
	// 5. parses matcher schema definitions
	return nil
}

// AddPolicy add policy
// storageKey: {namespace}\x00{schemaName}\x00{objectID}
func (m *mutation) AddPolicy(schemaName []byte, reader btuple.Reader) {
	// builds key values
	// updates secondary indexes
}

func (m *mutation) RemovePolicy(schemaName []byte, target btuple.Reader) (removed bool) {
	// search in secondary indexes
	// case1: found
	//	remove it from secondary indexes TODO: shall we?
	//	What if another txn tries to search the older value in secondary indexes
	//  which should be indexed.
	//  TODO: Or we should search the value in KV by the primary key retrieved from the secondary indexes

	//  remove value from kv
	// case2: not found, return false
	return false
}

func (m *mutation) UpdatePolicy(schemaName []byte, old, new btuple.Reader) (updated bool) {
	// search in secondary indexes
	// case1: found
	// 	TODO: Or we should search the value in KV by the primary key retrieved from the secondary indexes
	//	update value in kv
	// case2: not found, return false
	return false
}

func (m *mutation) Commit() error {
	// commit mutation txn
	// commit indexes changes
	return nil
}

func (m *mutation) Abort() error {
	// discard txn changes
	// discard indexes changes
	return nil
}
