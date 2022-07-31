package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func setupMockDB(t *testing.T, mockDb *mockDB) {
	sc := mockDb.NewTxnAt(1, true)
	checker := builderAsserter(mockDBInfo1)
	mockDb.CreateDB(t, sc, mockDBInfo1)
	err := sc.CommitTxn(context.TODO(), 2)
	assert.Nil(t, err)
	// waits all transactions finished
	assert.Nil(t, mockDb.WaitForMark(context.TODO(), 2))

	// check
	sc = mockDb.NewTxnAt(3, false)
	checker.Check(t, sc)
	err = sc.CommitTxn(context.TODO(), 3)
	assert.Nil(t, err)
}

func TestInsertExecutor(t *testing.T) {
	p := "./__test_tmp__/insert_exec"
	mockDb := OpenMockDB(t, p)
	defer func() {
		mockDb.Close()
		os.RemoveAll(p)
	}()
	setupMockDB(t, mockDb)

	sc := mockDb.NewTxnAt(4, true)

	executor, err := NewInsertExecutor(sc, plan.NewRawInsertPlan(mockDBDataSet, 1, 1), nil)
	assert.Nil(t, err)

	result, ids, err := Execute(executor, context.TODO())

	assert.Equal(t, len(result), len(mockDBDataSet))
	assert.Equal(t, len(ids), len(mockDBDataSet))

	expected := CloneTupleSet(mockDBDataSet)
	MergeDefaultValue(expected, mockDBInfo1.TableInfo[0])

	TuplesAsserter(t, expected, result)

}
