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
	"time"
)

type StartTransactionOption struct {
	timeout   time.Duration
	updateTxn bool
}

var DefaultStartTransactionOption = &StartTransactionOption{timeout: 60 * time.Second}

func (e *engine) StartTransaction(ctx context.Context, opt *StartTransactionOption) (*Session, error) {
	sessionId := e.sessionMgr.NewSession(e.newSessionCtx(ctx, opt.updateTxn), opt.timeout)
	return &Session{Id: sessionId}, nil
}

func (e *engine) AbortTransaction(ctx context.Context, session *Session) error {
	sessCtx, err := e.getSessionCtx(ctx, &BaseOptions{sessionId: &session.Id})
	if err != nil {
		return err
	}
	sessCtx.RollbackTxn(ctx)
	return nil
}

func (e *engine) CommitTransaction(ctx context.Context, session *Session) error {
	sessCtx, err := e.getSessionCtx(ctx, &BaseOptions{sessionId: &session.Id})
	if err != nil {
		return err
	}
	commitTs := e.orc.newCommitTs(sessCtx.GetTxn().ReadTs())
	return sessCtx.CommitTxn(ctx, commitTs)
}

func (e *engine) EndTransaction(ctx context.Context, session *Session) error {
	e.sessionMgr.CollectSessionCtx(ctx, session.Id)
	return nil
}
