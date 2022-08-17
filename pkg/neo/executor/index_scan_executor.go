package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/db"
	"github.com/casbin-mesh/neo/pkg/db/adapter"
	"github.com/casbin-mesh/neo/pkg/neo/codec"
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
	iterKey       []byte
}

func (i *indexScanExecutor) Init() {
	i.iter = i.GetTxn().NewIterator(adapter.DefaultIteratorOptions)
	i.iter.Seek(i.indexScanPlan.Prefix())
	if i.iter.ValidForPrefix(i.indexScanPlan.Prefix()) {
		i.iterKey = i.iter.Item().KeyCopy(nil)
	}
}

func (i *indexScanExecutor) Next(ctx context.Context, tuple *btuple.Modifier, rid *primitive.ObjectID) (next bool, err error) {
	if !i.iter.Valid() || !i.indexScanPlan.IsValid(i.iterKey) {
		return
	}

	if i.indexScanPlan.PrimaryIndex() {
		var rawVal []byte
		if rawVal, err = i.iter.Item().ValueCopy(nil); err != nil {
			return
		}
		if *rid, err = codec.ParseTupleRecordKeyFromPrimaryIndex(rawVal); err != nil {
			return
		}
	} else {
		if *rid, err = codec.ParseTupleRecordKeyFromSecondaryIndex(i.iterKey); err != nil {
			return
		}
	}

	if i.indexScanPlan.FetchTuple() {
		// TODO(weny): add second scan
	}

	i.iter.Next()
	i.iterKey = i.iter.Item().KeyCopy(nil)

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
