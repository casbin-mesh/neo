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

type insertExecutor struct {
	baseExecutor
	insertPlan    plan.InsertPlan
	childExecutor Executor
	iter          int
	tableInfo     *model.TableInfo
}

func (i *insertExecutor) Init() {
	if i.insertPlan.HasChildren() {
		i.childExecutor.Init()
	} else {
		i.iter = 0
	}
}

func (i *insertExecutor) Next(ctx context.Context, tuple *btuple.Modifier, rid *primitive.ObjectID) (next bool, err error) {
	if i.insertPlan.HasChildren() {
		if next, err = i.childExecutor.Next(ctx, tuple, rid); !next || err != nil { // occurs error or no more tuple
			return
		}
	} else {
		if i.iter == i.insertPlan.RawValuesSize() { // end
			return
		}
		curTuple := btuple.NewModifierFromBytes(
			codec.EncodeValues(i.insertPlan.RawValues()[i.iter]),
		)

		if err = curTuple.MergeDefaultValue(i.tableInfo); err != nil {
			return false, err
		}
		*tuple = curTuple
		i.iter++
	}

	*rid = primitive.NewObjectID()
	if err = i.GetTxn().Set(
		codec.TupleRecordKey(i.tableInfo.ID, *rid),
		btuple.NewTupleBuilder(
			btuple.SmallValueType, (*tuple).Values()...,
		).Encode(),
	); err != nil {
		return
	}

	// insert indices
	for _, index := range i.tableInfo.Indices {
		key, value := codec.IndexEntry(index, i.tableInfo.Columns, *tuple, *rid)
		if err = i.GetTxn().Set(key, value); err != nil {
			return false, err
		}
	}

	return true, nil
}

func NewInsertExecutor(ctx session.Context, insertPlan plan.InsertPlan, child Executor) (Executor, error) {
	dbInfo, err := ctx.GetCatalog().GetDBInfoByDBId(insertPlan.DBOid())
	if err != nil {
		return nil, err
	}
	tableInfo, err := dbInfo.TableById(insertPlan.TableOid())
	if err != nil {
		return nil, err
	}
	return &insertExecutor{
		baseExecutor:  newBaseExecutor(ctx),
		insertPlan:    insertPlan,
		childExecutor: child,
		tableInfo:     tableInfo,
	}, nil
}
