package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/db"
	"github.com/casbin-mesh/neo/pkg/db/adapter"
	"github.com/casbin-mesh/neo/pkg/neo/codec"
	"github.com/casbin-mesh/neo/pkg/neo/executor/expression"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type indexScanExecutor struct {
	baseExecutor
	indexScanPlan plan.IndexScanPlan
	tableInfo     *model.TableInfo
	iter          db.Iterator
}

func (i *indexScanExecutor) Init() {
	i.iter = i.GetTxn().NewIterator(adapter.DefaultIteratorOptions)
	i.iter.Seek(i.indexScanPlan.Prefix())
}

func (i *indexScanExecutor) Next(ctx context.Context, tuple *btuple.Modifier, rid *primitive.ObjectID) (next bool, err error) {
	if !i.iter.ValidForPrefix(i.indexScanPlan.Prefix()) {
		return
	}

	rawKey := i.iter.Item().KeyCopy(nil)
	if *rid, err = codec.ParseTupleRecordKeyFromSecondaryIndex(rawKey); err != nil {
		return
	}

	rawVal, err := i.iter.Item().ValueCopy(nil)
	if err != nil {
		return
	}

	tupleReader, err := btuple.NewReader(rawVal)
	if err != nil {
		return
	}

	*tuple = btuple.NewModifier(tupleReader.Values())

	i.iter.Next()
	predicate := i.indexScanPlan.Predicate()
	if predicate != nil {
		if res, err := predicate.Evaluate(i.GetSessionCtx(), i.indexScanPlan.GetEvalCtx(), *tuple, i.indexScanPlan.OutputSchema()); err == nil {
			if value, err := expression.TryGetBool(res); err != nil {
				return false, err
			} else if !value {
				return i.Next(ctx, tuple, rid)
			}
		} else {
			return false, err
		}
	}

	return true, nil
}

func (i *indexScanExecutor) Close() error {
	if i.iter != nil {
		i.iter.Close()
	}
	return nil
}

func NewIndexScanExecutor(ctx session.Context, scanPlan plan.IndexScanPlan) (Executor, error) {
	dbInfo, err := ctx.GetCatalog().GetDBInfoByDBId(scanPlan.DBOid())
	if err != nil {
		return nil, err
	}
	tableInfo, err := dbInfo.TableById(scanPlan.TableOid())
	if err != nil {
		return nil, err
	}
	return &indexScanExecutor{
		baseExecutor:  newBaseExecutor(ctx),
		indexScanPlan: scanPlan,
		tableInfo:     tableInfo,
	}, nil
}
