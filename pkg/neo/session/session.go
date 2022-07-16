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

package session

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/db"
	"github.com/casbin-mesh/neo/pkg/neo/meta"
	"github.com/casbin-mesh/neo/pkg/neo/schema"
	"github.com/dgraph-io/badger/v3/y"
)

type Context interface {
	CommitTxn(ctx context.Context, commitTs uint64) error
	RollbackTxn(ctx context.Context)
	GetMetaReaderWriter() meta.ReaderWriter
	GetSchemaReaderWriter() schema.ReaderWriter
	GetTxn() db.Txn
}

type ctx struct {
	txn     db.Txn
	meta    meta.ReaderWriter
	schema  schema.ReaderWriter
	txnMark *y.WaterMark
}

func (c ctx) GetSchemaReaderWriter() schema.ReaderWriter {
	return c.schema
}

func (c ctx) GetTxn() db.Txn {
	return c.txn
}

func (c ctx) GetMetaReaderWriter() meta.ReaderWriter {
	return c.meta
}

func (c ctx) CommitTxn(ctx context.Context, commitTs uint64) error {
	err := c.txn.CommitAt(commitTs, func(err error) {
		if err != nil {
			c.RollbackTxn(ctx)
			return
		}

		// TODO: handle failures
		err = c.meta.CommitAt(commitTs)
		err = c.schema.CommitAt(commitTs)
		c.txnMark.Done(commitTs)
	})
	return err
}

func (c ctx) RollbackTxn(ctx context.Context) {
	c.txn.Discard()
	c.meta.Rollback()
	c.schema.Rollback()
}

func NewSessionCtx(txn db.Txn, meta meta.ReaderWriter, schema schema.ReaderWriter, txnMark *y.WaterMark) Context {
	return &ctx{txn: txn, meta: meta, schema: schema, txnMark: txnMark}
}
