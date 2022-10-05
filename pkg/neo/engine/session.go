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
	"github.com/casbin-mesh/neo/pkg/db"
	"github.com/casbin-mesh/neo/pkg/neo/schema"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"sync"
	"time"
)

type Session struct {
	Id string
}

type SessionManager interface {
	CollectSessionCtx(ctx context.Context, sessionId string)
	AllocSessionCtx(ctx context.Context, txn db.Txn, writer schema.ReaderWriter, orc *oracle) session.Context
	NewSession(sessCtx session.Context, timeout time.Duration) string
	GetSession(sessionId string) (session.Context, bool)
}

type sessionManager struct {
	table sync.Map
	pool  *sync.Pool
}

func (s *sessionManager) CollectSessionCtx(ctx context.Context, sessionId string) {
	if v, ok := s.table.LoadAndDelete(sessionId); ok {
		if sessCtx, ok := v.(session.Context); ok {
			sessCtx.RollbackTxn(ctx)
			s.pool.Put(sessCtx)
		}
	}
}

func (s *sessionManager) NewSession(sessCtx session.Context, timeout time.Duration) string {
	sessionId := NewSessionId()
	s.table.Store(sessionId, sessCtx)

	go func() {
		select {
		case <-time.After(timeout):
			s.CollectSessionCtx(context.Background(), sessionId)
		}
	}()
	return sessionId
}

func (s *sessionManager) AllocSessionCtx(ctx context.Context, txn db.Txn, writer schema.ReaderWriter, orc *oracle) session.Context {
	sc := s.pool.Get().(session.Context)
	sc.Init(ctx, txn, writer, orc)
	return sc
}

func (s *sessionManager) GetSession(sessionId string) (session.Context, bool) {
	v, ok := s.table.Load(sessionId)
	if !ok {
		return nil, false
	}
	return v.(session.Context), true
}

func NewSessionManager() SessionManager {
	sessCtxPool := &sync.Pool{
		New: func() any {
			return session.Empty()
		},
	}
	return &sessionManager{
		table: sync.Map{},
		pool:  sessCtxPool,
	}
}
