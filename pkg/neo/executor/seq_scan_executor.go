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

type seqScanExecutor struct {
	baseExecutor
	seqScanPlan plan.SeqScanPlan
	tableInfo   *model.TableInfo
	iter        db.Iterator
	prefix      []byte
}

func (s *seqScanExecutor) Init() {
	s.prefix = codec.TupleRecordBegin(s.tableInfo.ID)
	s.iter = s.GetTxn().NewIterator(adapter.DefaultIteratorOptions)
	s.iter.Seek(s.prefix)
}

func (s *seqScanExecutor) Close() error {
	if s.iter != nil {
		s.iter.Close()
	}
	return nil
}

func (s *seqScanExecutor) Next(ctx context.Context, tuple *btuple.Modifier, rid *primitive.ObjectID) (next bool, err error) {
	if !s.iter.ValidForPrefix(s.prefix) {
		return
	}

	rawKey := s.iter.Item().KeyCopy(nil)
	if *rid, err = codec.ParseTupleRecordKey(rawKey); err != nil {
		return
	}

	rawVal, err := s.iter.Item().ValueCopy(nil)
	if err != nil {
		return
	}

	// TODO(weny): create modifier directly
	tupleReader, err := btuple.NewReader(rawVal)
	if err != nil {
		return
	}

	//TODO:(weny): generates tuple following the output schema
	*tuple = btuple.NewModifier(tupleReader.Values())

	s.iter.Next()

	predicate := s.seqScanPlan.Predicate()
	if predicate != nil {
		if res, err := predicate.Evaluate(s.GetSessionCtx(), s.seqScanPlan.GetEvalCtx(), *tuple, s.seqScanPlan.OutputSchema()); err == nil {
			if value, err := expression.TryGetBool(res); err != nil {
				return false, err
			} else if !value {
				return s.Next(ctx, tuple, rid)
			}
		} else {
			return false, err
		}
	}

	return true, nil
}

func NewSeqScanExecutor(ctx session.Context, scanPlan plan.SeqScanPlan) (Executor, error) {
	dbInfo, err := ctx.GetCatalog().GetDBInfoByDBId(scanPlan.DBOid())
	if err != nil {
		return nil, err
	}
	tableInfo, err := dbInfo.TableById(scanPlan.TableOid())
	if err != nil {
		return nil, err
	}
	return &seqScanExecutor{
		baseExecutor: newBaseExecutor(ctx),
		seqScanPlan:  scanPlan,
		tableInfo:    tableInfo,
	}, nil
}
