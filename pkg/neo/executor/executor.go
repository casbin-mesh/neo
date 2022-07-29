package executor

import (
	"github.com/casbin-mesh/neo/pkg/db"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type Executor interface {
	Init()
	Next(tuple *btuple.Reader, rid *primitive.ObjectID) (bool, error)
}

type baseExecutor struct {
	ctx session.Context
}

func (b *baseExecutor) GetSessionCtx() session.Context {
	return b.ctx
}

func (b *baseExecutor) GetTxn() db.Txn {
	return b.ctx.GetTxn()
}

func newBaseExecutor(ctx session.Context) baseExecutor {
	return baseExecutor{ctx: ctx}
}
