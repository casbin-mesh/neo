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

package engine

import (
	"context"
	"errors"
	"github.com/casbin-mesh/neo/pkg/db"
	badgerAdapter "github.com/casbin-mesh/neo/pkg/db/adapter/badger"
	"github.com/casbin-mesh/neo/pkg/neo/executor"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/index"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	casbinModel "github.com/casbin-mesh/neo/pkg/neo/optimizer/model"
	"github.com/casbin-mesh/neo/pkg/neo/schema"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/dgraph-io/badger/v3"
)

type M = map[string]interface{}
type A = []interface{}

type Table interface {
	InsertOne(ctx context.Context, data A, opts ...*InsertOptions) (A, error)
	InsertMany(ctx context.Context, data []A, opts ...*InsertOptions) ([]A, error)
	UpdateOne(ctx context.Context, data A, update A, opts ...*UpdateOptions) (A, error)
	UpdateMany(ctx context.Context, data []A, update []A, opts ...*UpdateOptions) ([]A, error)
	DeleteOne(ctx context.Context, data A, opts ...*DeleteOptions) (A, error)
	DeleteMany(ctx context.Context, data []A, opts ...*DeleteOptions) ([]A, error)
	EnforceOne(ctx context.Context, data A, opts ...*EnforceOptions) (bool, error)
	EnforceMany(ctx context.Context, data []A, opts ...*EnforceOptions) ([]bool, error)
	FindOne(ctx context.Context, filter interface{}) (M, error)
	Find(ctx context.Context, filter interface{}, opts ...*FindOptions) ([]M, error)
}

type Namespace interface {
	Table(name string) Table
}

type Engine interface {
	AddNamespaceFromString(ctx context.Context, namespace string, rawModel string, opts ...*BaseOptions) error

	StartTransaction(ctx context.Context, opt *StartTransactionOption) (*Session, error)
	AbortTransaction(ctx context.Context, session *Session) error
	CommitTransaction(ctx context.Context, session *Session) error
	EndTransaction(ctx context.Context, session *Session) error

	Namespace(ns string) Namespace

	Close() error

	getSessionCtx(ctx context.Context, bo *BaseOptions) (session.Context, error)
	commitSession(ctx context.Context, sessCtx session.Context) error
	discardSession(ctx context.Context, sessCtx session.Context)
}

var (
	ErrSessionTimeout = errors.New("session timeout")
)

type engine struct {
	db         db.DB
	schema     index.Index[*model.DBInfo]
	orc        *oracle
	sessionMgr SessionManager
}

func (e *engine) Namespace(ns string) Namespace {
	return &namespace{
		dbName: ns,
		engine: e,
	}
}

func (e *engine) Close() error {
	e.orc.Stop()
	return nil
}

func (e *engine) newSessionCtx(ctx context.Context) session.Context {
	readTs := e.orc.readTs()
	txn := e.db.NewTransactionAt(readTs, true)
	schemaTxn := e.schema.NewTransactionAt(readTs, true)
	sessionCtx := e.sessionMgr.AllocSessionCtx(ctx, txn, schema.New(schemaTxn), e.orc)
	return sessionCtx
}

func (e *engine) commitSession(ctx context.Context, sessCtx session.Context) error {
	commitTs := e.orc.newCommitTs(sessCtx.GetTxn().ReadTs())
	return sessCtx.CommitTxn(ctx, commitTs)
}

func (e *engine) discardSession(ctx context.Context, sessCtx session.Context) {
	sessCtx.RollbackTxn(ctx)
}

func (e *engine) getSessionCtx(ctx context.Context, bo *BaseOptions) (session.Context, error) {
	if bo.sessionId != nil {
		_, ok := e.sessionMgr.GetSession(*bo.sessionId)
		if !ok {
			return nil, ErrSessionTimeout
		}
	}
	return e.newSessionCtx(ctx), nil
}

func (e *engine) AddNamespaceFromString(ctx context.Context, namespace string, rawModel string, opts ...*BaseOptions) error {
	g, err := casbinModel.NewGeneratorFromString(rawModel)
	if err != nil {
		return err
	}
	dbInfo := g.GenerateDB(namespace)

	opt := MergeBaseOptions(opts...)

	sessCtx, err := e.getSessionCtx(ctx, opt)
	defer e.discardSession(ctx, sessCtx)
	if err != nil {
		return err
	}

	builder := executor.NewExecutorBuilder(sessCtx)
	exec, err := builder.Build(plan.NewCreateDBPlan(dbInfo)), builder.Error()
	if err != nil {
		return err
	}
	_, _, err = executor.Execute(exec, ctx)
	if err != nil {
		return err
	}

	if opt.AutoCommit() {
		err = e.commitSession(ctx, sessCtx)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewEngineFromPath(path string) (Engine, error) {
	// open database
	instance, err := badgerAdapter.OpenManaged(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	schemaIndex := index.New[*model.DBInfo](index.Options{})
	// TODO: restore schema from db

	orc := newOracle()
	orc.setDiscard = instance.SetDiscardTs
	orc.nextTxnTs = instance.MaxVersion()
	orc.txnMark.Done(orc.nextTxnTs)
	orc.readMark.Done(orc.nextTxnTs)
	orc.incrementNextTs()

	return &engine{
		db:         instance,
		orc:        orc,
		schema:     schemaIndex,
		sessionMgr: NewSessionManager(),
	}, nil
}
