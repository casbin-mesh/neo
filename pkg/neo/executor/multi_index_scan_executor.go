package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
	"github.com/casbin-mesh/neo/pkg/primitive/utils"
)

type multiIndexScanExecutor struct {
	baseExecutor
	multiIndexScanPlan plan.MultiIndexScan
	tableInfo          *model.TableInfo
	// left executor yielded set smaller than right one
	left     Executor
	right    Executor
	prepared bool
	// TODO: uses a container interface instead.
	// TODO: compares performance to the other implementations, e.g., using adaptive radix tree as the container
	hashMap map[primitive.ObjectID]btuple.Modifier
}

func (m *multiIndexScanExecutor) fetchAndBuildHashTable(ctx context.Context) (err error) {
	m.left.Init()
	m.right.Init()
	var next bool
	for {
		var (
			tuple btuple.Modifier
			rid   primitive.ObjectID
		)
		if next, err = m.left.Next(ctx, &tuple, &rid); err != nil {
			return
		}
		if !next {
			break
		}
		if !rid.IsEmpty() {
			m.hashMap[rid] = tuple
		}
	}
	return m.left.Close()
}

func (m *multiIndexScanExecutor) Init() {
	m.hashMap = make(map[primitive.ObjectID]btuple.Modifier)
}

func (m *multiIndexScanExecutor) probe(ctx context.Context, tuple *btuple.Modifier, rid *primitive.ObjectID) (next bool, err error) {
	var (
		right btuple.Modifier
	)
	for {
		if next, err = m.right.Next(ctx, &right, rid); err != nil {
			return
		}
		if !next {
			return
		}
		if _, ok := m.hashMap[*rid]; ok {
			mo, _ := utils.MergeModifier(
				m.hashMap[*rid],
				m.multiIndexScanPlan.LeftOutputSchema(),
				right,
				m.multiIndexScanPlan.RightOutputSchema(),
			)
			*tuple = mo
			return true, nil
		}
	}
}

func (m *multiIndexScanExecutor) Next(ctx context.Context, tuple *btuple.Modifier, rid *primitive.ObjectID) (next bool, err error) {
	if !m.prepared {
		if err = m.fetchAndBuildHashTable(ctx); err != nil {
			return false, err
		}
		m.prepared = true
	}
	// TODO: parallel probing
	return m.probe(ctx, tuple, rid)
}

func (m *multiIndexScanExecutor) Close() error {
	return m.right.Close()
}

func NewMultiIndexScanExecutor(ctx session.Context, scanPlan plan.MultiIndexScan, left, right Executor) (Executor, error) {
	dbInfo, err := ctx.GetCatalog().GetDBInfoByDBId(scanPlan.DBOid())
	if err != nil {
		return nil, err
	}
	tableInfo, err := dbInfo.TableById(scanPlan.TableOid())
	if err != nil {
		return nil, err
	}
	return &multiIndexScanExecutor{
		baseExecutor:       newBaseExecutor(ctx),
		multiIndexScanPlan: scanPlan,
		tableInfo:          tableInfo,
		left:               left,
		right:              right,
	}, nil
}
