package executor

import (
	"errors"
	"fmt"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/session"
)

type executorBuilder struct {
	ctx session.Context
	err error
}

func (b *executorBuilder) ResetError() {
	b.err = nil
}

func (b *executorBuilder) Error() error {
	return b.err
}

func NewExecutorBuilder(ctx session.Context) *executorBuilder {
	return &executorBuilder{
		ctx: ctx,
	}
}

func (b *executorBuilder) Build(p plan.AbstractPlan) Executor {
	return b.build(p)
}

func (b *executorBuilder) build(p plan.AbstractPlan) Executor {
	switch v := p.(type) {
	case plan.InsertPlan:
		return b.buildInsertPlan(v)
	case plan.UpdatePlan:
		return b.buildUpdatePlan(v)
	case plan.IndexScanPlan:
		return b.buildIndexScanPlan(v)
	case plan.SeqScanPlan:
		return b.buildSeqScanPlan(v)
	case plan.DeletePlan:
		return b.buildDeletePlan(v)
	case plan.LimitPlan:
		return b.buildLimitPlan(v)
	case plan.SchemaPlan:
		return b.buildSchemaPlan(v)
	case plan.MultiIndexScan:
		return b.buildMultiIndexScan(v)
	default:
		b.err = fmt.Errorf("unknown Plan %T", p)
		return nil
	}
}

var (
	ErrMissChildPlan = errors.New("miss child plan")
)

func (b *executorBuilder) catchErr(err error) bool {
	if err != nil {
		b.err = err
	}
	return err != nil
}

func (b *executorBuilder) buildUpdatePlan(p plan.UpdatePlan) Executor {
	if !p.HasChildren() {
		b.catchErr(ErrMissChildPlan)
		return nil
	}
	childExec := b.build(p.GetChildAt(0))
	exec, err := NewUpdateExecutor(b.ctx, p, childExec)
	if b.catchErr(err) {
		return nil
	}
	return exec
}

func (b *executorBuilder) buildSchemaPlan(p plan.SchemaPlan) Executor {
	return NewSchemaExec(b.ctx, p)
}

func (b *executorBuilder) buildDeletePlan(p plan.DeletePlan) Executor {
	if !p.HasChildren() {
		b.catchErr(ErrMissChildPlan)
		return nil
	}
	childExec := b.build(p.GetChildAt(0))
	exec, err := NewDeleteExecutor(b.ctx, p, childExec)
	if b.catchErr(err) {
		return nil
	}
	return exec
}

func (b *executorBuilder) buildLimitPlan(p plan.LimitPlan) Executor {
	if !p.HasChildren() {
		b.catchErr(ErrMissChildPlan)
		return nil
	}
	childExec := b.build(p.GetChildAt(0))
	return NewLimitExecutor(b.ctx, p, childExec)
}

func (b *executorBuilder) buildSeqScanPlan(p plan.SeqScanPlan) Executor {
	exec, err := NewSeqScanExecutor(b.ctx, p)
	if b.catchErr(err) {
		return nil
	}
	return exec
}

func (b *executorBuilder) buildInsertPlan(p plan.InsertPlan) Executor {
	if p.HasChildren() {
		childExec := b.build(p.GetChildAt(0))
		exec, err := NewInsertExecutor(b.ctx, p, childExec)
		if b.catchErr(err) {
			return nil
		}
		return exec
	}
	exec, err := NewInsertExecutor(b.ctx, p, nil)
	if b.catchErr(err) {
		return nil
	}
	return exec
}

func (b *executorBuilder) buildIndexScanPlan(v plan.IndexScanPlan) Executor {
	exec, err := NewIndexScanExecutor(b.ctx, v)
	if b.catchErr(err) {
		return nil
	}
	return exec
}

func (b *executorBuilder) buildMultiIndexScan(v plan.MultiIndexScan) Executor {
	if len(v.GetChildren()) != 2 {
		b.catchErr(ErrMissChildPlan)
		return nil
	}
	leftExec, err := b.build(v.GetChildAt(0)), b.err
	if err != nil {
		return nil
	}
	rightExec, err := b.build(v.GetChildAt(1)), b.err
	if err != nil {
		return nil
	}
	exec, err := NewMultiIndexScanExecutor(b.ctx, v, leftExec, rightExec)
	if b.catchErr(err) {
		return nil
	}
	return exec
}
