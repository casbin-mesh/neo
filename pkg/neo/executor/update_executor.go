package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/neo/codec"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type updateExecutor struct {
	baseExecutor
	updatePlan    plan.UpdatePlan
	childExecutor Executor
	tableInfo     *model.TableInfo
}

func (u *updateExecutor) Init() {
	u.childExecutor.Init()
}

func (u *updateExecutor) Close() error {
	return u.childExecutor.Close()
}

func (u *updateExecutor) Next(ctx context.Context, tuple *btuple.Modifier, rid *primitive.ObjectID) (next bool, err error) {
	cond := func() bool {
		next, err = u.childExecutor.Next(ctx, tuple, rid)
		return next && err == nil
	}

	for cond() {
		u.GenerateUpdateTuple(tuple)

		if err = u.GetTxn().Set(
			codec.TupleRecordKey(u.tableInfo.ID, *rid),
			btuple.NewTupleBuilder(
				btuple.SmallValueType, (*tuple).Values()...,
			).Encode(),
		); err != nil {
			return
		}
		// TODO: update index info
	}

	return
}

func (u *updateExecutor) GenerateUpdateTuple(b *btuple.Modifier) {
	updateAttrs := u.updatePlan.GetUpdateAttrs()
	for i, _ := range u.tableInfo.Columns {
		if modifier, ok := updateAttrs[i]; ok {
			switch modifier.Type() {
			case plan.ModifierSet:
				(*b).Set(i, modifier.Value().([]byte))
			}
		}
	}
}

func NewUpdateExecutor(ctx session.Context, updatePlan plan.UpdatePlan, child Executor) (Executor, error) {
	dbInfo, err := ctx.GetCatalog().GetDBInfoByDBId(updatePlan.DBOid())
	if err != nil {
		return nil, err
	}
	tableInfo, err := dbInfo.TableById(updatePlan.TableOid())
	if err != nil {
		return nil, err
	}
	return &updateExecutor{
		baseExecutor:  newBaseExecutor(ctx),
		updatePlan:    updatePlan,
		childExecutor: child,
		tableInfo:     tableInfo,
	}, nil
}
