package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/db"
	"github.com/casbin-mesh/neo/pkg/db/adapter"
	"github.com/casbin-mesh/neo/pkg/neo/codec"
	"github.com/casbin-mesh/neo/pkg/neo/executor/expression"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type tableRowIdScanExecutor struct {
	baseExecutor
	plan  *plan.TableRowIdScan
	iter  db.Iterator
	child Executor
}

func (t *tableRowIdScanExecutor) Init() {
	t.child.Init()
	t.iter = t.GetTxn().NewIterator(adapter.DefaultIteratorOptions)
}

func (t *tableRowIdScanExecutor) Next(ctx context.Context, tuple *btuple.Modifier, rid *primitive.ObjectID) (next bool, err error) {

	for {
		if next, err = t.child.Next(ctx, tuple, rid); !next || err != nil {
			return
		}
		key := codec.TupleRecordKey(t.plan.TableOid(), *rid)
		t.iter.Seek(key)

		if !t.iter.Valid() {
			return
		}

		rawVal, err := t.iter.Item().ValueCopy(nil)
		if err != nil {
			return false, err
		}

		// TODO: create modifier directly
		tupleReader, err := btuple.NewReader(rawVal)
		if err != nil {
			return false, err
		}

		//TODO: generates tuple following the output schema
		*tuple = btuple.NewModifier(tupleReader.Values())

		predicate := t.plan.Predicate()
		if predicate != nil {
			if res, err := predicate.Evaluate(t.GetSessionCtx(), t.plan.GetEvalCtx(), *tuple, t.plan.OutputSchema()); err == nil {
				if value, err := expression.TryGetBool(res); err != nil {
					return false, err
				} else if value {
					break
				}
			} else {
				return false, err
			}
		} else {
			break
		}

	}
	return true, nil
}

func (t *tableRowIdScanExecutor) Close() error {
	if t.iter != nil {
		t.iter.Close()
	}
	err := t.child.Close()
	if err != nil {
		return err
	}
	return nil
}

func NewTableRowIdScanExecutor(ctx session.Context, plan *plan.TableRowIdScan, child Executor) (Executor, error) {
	return &tableRowIdScanExecutor{
		baseExecutor: newBaseExecutor(ctx),
		plan:         plan,
		child:        child,
	}, nil
}
