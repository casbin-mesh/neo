package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type limitExecutor struct {
	baseExecutor
	limitPlan     plan.LimitPlan
	childExecutor Executor
	count         int
	finished      bool
}

func (l *limitExecutor) Next(ctx context.Context, tuple *btuple.Modifier, rid *primitive.ObjectID) (next bool, err error) {
	if l.count == l.limitPlan.Limit() || l.finished {
		return
	}
	next, err = l.childExecutor.Next(ctx, tuple, rid)
	if !next || err != nil {
		l.finished = true
		return
	}

	l.count++
	return true, nil
}

func (l *limitExecutor) Init() {
	l.childExecutor.Init()
}

func (l *limitExecutor) Close() error {
	return l.childExecutor.Close()
}

func NewLimitExecutor(ctx session.Context, limitPlan plan.LimitPlan, child Executor) Executor {
	return &limitExecutor{
		baseExecutor:  newBaseExecutor(ctx),
		limitPlan:     limitPlan,
		childExecutor: child,
	}
}
