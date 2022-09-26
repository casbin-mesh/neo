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
	"github.com/casbin-mesh/neo/pkg/neo/catalog"
	"github.com/casbin-mesh/neo/pkg/neo/meta"
	"github.com/casbin-mesh/neo/pkg/neo/schema"
	"github.com/dgraph-io/badger/v3/y"
	"sync"
)

type Oracle interface {
	DoneRead(readTs uint64)
	IncNextTs()
	DoneCommit(commitTs uint64)
}

type Context interface {
	Init(ctx2 context.Context, txn db.Txn, schema schema.ReaderWriter, oracle Oracle)
	CommitTxn(ctx context.Context, commitTs uint64) error
	RollbackTxn(ctx context.Context)
	GetCatalog() catalog.Catalog
	GetMetaReaderWriter() meta.ReaderWriter
	GetSchemaReaderWriter() schema.ReaderWriter
	GetTxn() db.Txn
}

type ctx struct {
	sync.Mutex
	discarded bool
	committed bool
	context.Context
	txn     db.Txn
	catalog catalog.Catalog
	meta    meta.ReaderWriter
	schema  schema.ReaderWriter
	txnMark *y.WaterMark
	oracle  Oracle
}

func (c *ctx) Init(ctx2 context.Context, txn db.Txn, schema schema.ReaderWriter, oracle Oracle) {
	c.discarded = false
	c.committed = false
	c.Context = ctx2
	c.txn = txn
	c.schema = schema
	c.oracle = oracle
	c.catalog = catalog.NewCatalog(schema, txn)
}

func (c *ctx) GetSchemaReaderWriter() schema.ReaderWriter {
	return c.schema
}

func (c *ctx) GetCatalog() catalog.Catalog {
	return c.catalog
}

func (c *ctx) GetTxn() db.Txn {
	return c.txn
}

func (c *ctx) GetMetaReaderWriter() meta.ReaderWriter {
	if c.meta != nil {
		return c.meta
	}
	return c.catalog.GetMetaRW()
}

func (c *ctx) CommitTxn(ctx context.Context, commitTs uint64) error {
	c.Lock()
	defer c.Unlock()
	if c.committed {
		return nil
	}
	c.committed = true
	err := c.txn.CommitAt(commitTs, func(err error) {
		if err != nil {
			c.RollbackTxn(ctx)

			c.schema.Rollback()

			if c.oracle != nil {
				c.oracle.DoneCommit(commitTs)
			}
			return
		}

		if c.meta != nil {
			err = c.meta.CommitAt(commitTs)
		}
		if c.txnMark != nil {
			c.txnMark.Done(commitTs)
		}
		if c.oracle != nil {
			c.oracle.DoneCommit(commitTs)
		}
		c.schema.CommitAt(commitTs)
	})
	return err
}

func (c *ctx) RollbackTxn(ctx context.Context) {
	c.Lock()
	defer c.Unlock()

	if c.committed || c.discarded {
		return
	}
	c.discarded = true

	if c.oracle != nil {
		c.oracle.DoneRead(c.txn.ReadTs())
	}

	c.txn.Discard()
	c.schema.Rollback()

	if c.meta != nil {
		c.meta.Rollback()
	}

}

// NewSessionManually used in test
func NewSessionManually(txn db.Txn, meta meta.ReaderWriter, schema schema.ReaderWriter, txnMark *y.WaterMark) Context {
	sessCtx := &ctx{
		txn:     txn,
		catalog: catalog.NewCatalogWithMetaRW(meta, schema, txn),
		meta:    meta,
		schema:  schema,
		txnMark: txnMark,
	}
	return sessCtx
}

func Empty() Context {
	return &ctx{}
}

func NewSessionCtx(ctx2 context.Context, txn db.Txn, schema schema.ReaderWriter, oracle Oracle) Context {
	sessCtx := &ctx{
		Context: ctx2,
		txn:     txn,
		catalog: catalog.NewCatalog(schema, txn),
		schema:  schema,
		oracle:  oracle,
	}
	return sessCtx
}
