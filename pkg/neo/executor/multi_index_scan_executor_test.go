package executor

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/casbin-mesh/neo/pkg/expression"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/parser"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewMultiIndexScanExecutor(t *testing.T) {
	p := "./__test_tmp__/multi_index_scan_exec"
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

	// multi-index scan
	// where subject = bob and object = data2
	sc = mockDb.NewTxnAt(6, true)
	builder := executorBuilder{ctx: sc}

	// subject index scan
	idxId := uint64(1)
	var indexId [8]byte
	binary.BigEndian.PutUint64(indexId[:], idxId)
	// scan from
	indexPrefix := []byte(fmt.Sprintf("i%s_%s", indexId, "bob"))

	mockExpr, accessor := expression.NewExpression(parser.MustParseFromString("p.subject == \"bob\""))
	ctx := ast.NewContext()
	ctx.AddAccessor("p", accessor)

	sujIndexScanPlan := plan.NewIndexScanPlan(model.NewIndexSchemaReader(mockDBInfo1.TableInfo[0], 0), indexPrefix, mockExpr, ctx, 1, 1)

	// object index scan
	idxId = uint64(2)
	binary.BigEndian.PutUint64(indexId[:], idxId)
	// scan from
	objIndexPrefix := []byte(fmt.Sprintf("i%s_%s", indexId, "data2"))

	mockExpr2, accessor2 := expression.NewExpression(parser.MustParseFromString("p.object == \"data2\""))
	ctx2 := ast.NewContext()
	ctx2.AddAccessor("p", accessor2)

	objIndexScanPlan := plan.NewIndexScanPlan(model.NewIndexSchemaReader(mockDBInfo1.TableInfo[0], 1), objIndexPrefix, mockExpr2, ctx2, 1, 1)

	// multi scan node
	multiIndexScanPlan := plan.NewMultiIndexScan([]plan.AbstractPlan{sujIndexScanPlan, objIndexScanPlan}, 1, 1)
	exec, err := builder.Build(multiIndexScanPlan), builder.Error()
	assert.Nil(t, err)

	// execute
	_, ids, err := Execute(exec, context.TODO())
	assert.Nil(t, err)
	assert.Nil(t, sc.CommitTxn(context.TODO(), 7))

	// generate expected ids
	var expected []primitive.ObjectID
	for i, modifier := range inserted {
		if bytes.Compare(modifier.ValueAt(0).Clone(), []byte("bob")) == 0 && bytes.Compare(modifier.ValueAt(1).Clone(), []byte("data2")) == 0 {
			expected = append(expected, insertedIds[i])
		}
	}

	IdsAsserter(t, expected, ids)
}
