package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewLimitExecutor(t *testing.T) {
	p := "./__test_tmp__/limit_exec"
	mockDb := OpenMockDB(t, p)
	defer func() {
		mockDb.Close()
		os.RemoveAll(p)
	}()
	setupMockDB(t, mockDb)

	// insert tuples
	sc := mockDb.NewTxnAt(4, true)
	inserted, insertedIds, err := mockDb.InsertTuples(t, sc, 1, 1, mockDBDataSet)
	err = sc.CommitTxn(context.TODO(), 5)
	assert.Nil(t, err)
	assert.Nil(t, mockDb.WaitForMark(context.TODO(), 5))

	sc = mockDb.NewTxnAt(6, false)
	builder := executorBuilder{ctx: sc}
	limit, err := builder.Build(
		plan.NewLimitPlan([]plan.AbstractPlan{
			plan.NewSeqScanPlan(mockDBInfo1.TableInfo[0], nil, nil, 1, 1),
		}, 10),
	), builder.Error()

	result, ids, err := Execute(limit, context.TODO())
	assert.Nil(t, nil)

	IdsAsserter(t, insertedIds, ids)
	TuplesAsserter(t, inserted, result)

	scan, err := NewSeqScanExecutor(sc, plan.NewSeqScanPlan(mockDBInfo1.TableInfo[0], nil, nil, 1, 1))
	assert.Nil(t, nil)

	limit = NewLimitExecutor(sc, plan.NewLimitPlan(nil, 1), scan)
	result, ids, err = Execute(limit, context.TODO())
	assert.Nil(t, nil)

	IdsAsserter(t, insertedIds[:1], ids)
	TuplesAsserter(t, inserted[:1], result)

	err = sc.CommitTxn(context.TODO(), 7)
	assert.Nil(t, mockDb.WaitForMark(context.TODO(), 7))
}
