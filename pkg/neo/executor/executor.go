package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/db"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type Executor interface {
	Init()
	Next(ctx context.Context, tuple *btuple.Modifier, rid *primitive.ObjectID) (bool, error)
	Close() error
}

func (b *baseExecutor) Close() error {
	return nil
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

func Execute(executor Executor, ctx context.Context) (result []btuple.Modifier, ids []primitive.ObjectID, err error) {
	executor.Init()
	var (
		next bool
	)
	for {
		var (
			tuple btuple.Modifier
			rid   primitive.ObjectID
		)
		if next, err = executor.Next(ctx, &tuple, &rid); err != nil {
			return
		}
		if !next {
			break
		}
		if tuple != nil {
			result = append(result, tuple)
			ids = append(ids, rid)
		}

	}
	err = executor.Close()
	return
}
