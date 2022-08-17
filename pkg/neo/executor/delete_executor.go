package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/db"
	"github.com/casbin-mesh/neo/pkg/neo/codec"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type deleteExecutor struct {
	baseExecutor
	deletePlan    plan.DeletePlan
	childExecutor Executor
	tableInfo     *model.TableInfo
}

func (d *deleteExecutor) Init() {
	d.childExecutor.Init()
}

func (d *deleteExecutor) Close() error {
	return d.childExecutor.Close()
}

func (d *deleteExecutor) Next(ctx context.Context, tuple *btuple.Modifier, rid *primitive.ObjectID) (next bool, err error) {
	cond := func() bool {
		next, err = d.childExecutor.Next(ctx, tuple, rid)
		return next && err == nil
	}

	for cond() {
		if err = d.GetTxn().Delete(codec.TupleRecordKey(d.tableInfo.ID, *rid)); err != nil {
			return false, err
		}
	}

	// delete index info
	for _, index := range d.tableInfo.Indices {
		if err = codec.IndexEntries(index, *tuple, *rid, func(key, value []byte) error {
			e := d.GetTxn().Delete(key)
			if e == db.ErrKeyNotFound {
				return nil
			}
			return e
		}); err != nil {
			return false, err
		}
	}

	return
}

func NewDeleteExecutor(ctx session.Context, deletePlan plan.DeletePlan, child Executor) (Executor, error) {
	dbInfo, err := ctx.GetCatalog().GetDBInfoByDBId(deletePlan.DbOid())
	if err != nil {
		return nil, err
	}
	tableInfo, err := dbInfo.TableById(deletePlan.TableOid())
	if err != nil {
		return nil, err
	}
	return &deleteExecutor{
		baseExecutor:  newBaseExecutor(ctx),
		deletePlan:    deletePlan,
		childExecutor: child,
		tableInfo:     tableInfo,
	}, nil
}
