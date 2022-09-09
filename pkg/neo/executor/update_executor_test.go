package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/primitive/value"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewUpdateExecutor(t *testing.T) {
	p := "./__test_tmp__/update_exec"
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

	// update tuples
	sc = mockDb.NewTxnAt(5, true)
	scan, err := NewSeqScanExecutor(sc, plan.NewSeqScanPlan(mockDBInfo1.TableInfo[0], nil, nil, 1, 1))
	assert.Nil(t, err)

	updateAttrs := map[int]plan.Modifier{}
	updateAttrs[3] = plan.NewModifier(plan.ModifierSet, value.NewStringValue("deny"))

	builder := executorBuilder{ctx: sc}
	exec, err := builder.Build(
		plan.NewUpdatePlan(
			[]plan.AbstractPlan{
				plan.NewSeqScanPlan(mockDBInfo1.TableInfo[0], nil, nil, 1, 1),
			},
			1, 1, updateAttrs),
	), builder.Error()

	assert.Nil(t, err)

	result, ids, err := Execute(exec, context.TODO())

	assert.Nil(t, sc.CommitTxn(context.TODO(), 6))
	assert.Nil(t, err)

	// update should not generate result
	assert.Nil(t, result)
	assert.Nil(t, ids)

	// wait for commit
	assert.Nil(t, mockDb.WaitForMark(context.TODO(), 6))

	// verify result after update
	sc = mockDb.NewTxnAt(7, true)
	scan, err = NewSeqScanExecutor(sc, plan.NewSeqScanPlan(mockDBInfo1.TableInfo[0], nil, nil, 1, 1))
	assert.Nil(t, err)
	result, ids, err = Execute(scan, context.TODO())
	assert.Nil(t, sc.CommitTxn(context.TODO(), 8))

	// generate expected set
	expected := CloneTupleSet(inserted)
	UpdateValue(expected, mockDBInfo1.TableInfo[0], updateAttrs)

	IdsAsserter(t, insertedIds, ids)
	TuplesAsserter(t, expected, result)

}
