package executor

import (
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

func (i *insertExecutor) Next(tuple *btuple.Modifier, rid *primitive.ObjectID) (next bool, err error) {
	if i.insertPlan.HasChildren() {
		if next, err = i.childExecutor.Next(tuple, rid); !next || err != nil { // occurs error or no more tuple
			return
		}
	} else {
		if i.iter == i.insertPlan.RawValuesSize()-1 { // end
			return
		}
		*tuple = i.insertPlan.RawValues()[i.iter]
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

	// TODO: update index info

	return true, nil
}

func NewInsertExecutor(ctx session.Context) Executor {
	return &insertExecutor{
		baseExecutor: newBaseExecutor(ctx),
	}
}
