package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewDeleteExecutor(t *testing.T) {
	p := "./__test_tmp__/delete_exec"
	mockDb := OpenMockDB(t, p)
	defer func() {
		mockDb.Close()
		os.RemoveAll(p)
	}()
	setupMockDB(t, mockDb)

	// insert tuples
	sc := mockDb.NewTxnAt(4, true)
	_, _, err := mockDb.InsertTuples(t, sc, 1, 1, mockDBDataSet)
	err = sc.CommitTxn(context.TODO(), 5)
	assert.Nil(t, err)

	assert.Nil(t, mockDb.WaitForMark(context.TODO(), 5))

	// delete tuples
	sc = mockDb.NewTxnAt(6, true)
	builder := executorBuilder{ctx: sc}
	exec, err := builder.Build(
		plan.NewDeletePlan(
			[]plan.AbstractPlan{
				plan.NewSeqScanPlan(mockDBInfo1.TableInfo[0], nil, 1, 1),
			},
			1, 1),
	), builder.Error()
	assert.Nil(t, err)
	result, ids, err := Execute(exec, context.TODO())
	assert.Nil(t, err)
	err = sc.CommitTxn(context.TODO(), 7)
	assert.Nil(t, err)

	assert.Nil(t, mockDb.WaitForMark(context.TODO(), 7))

	// delete should not generate result
	assert.Nil(t, result)
	assert.Nil(t, ids)

	// scan again
	sc = mockDb.NewTxnAt(8, true)
	scan, err := NewSeqScanExecutor(sc, plan.NewSeqScanPlan(mockDBInfo1.TableInfo[0], nil, 1, 1))
	result, ids, err = Execute(scan, context.TODO())
	err = sc.CommitTxn(context.TODO(), 9)
	assert.Nil(t, err)

	assert.Nil(t, mockDb.WaitForMark(context.TODO(), 9))

	assert.Nil(t, result)
	assert.Nil(t, ids)
}
