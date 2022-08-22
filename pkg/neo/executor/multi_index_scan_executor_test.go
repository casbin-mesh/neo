package executor

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/casbin-mesh/neo/pkg/neo/executor/expression"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
	"github.com/casbin-mesh/neo/pkg/primitive/value"
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

	mockExpr := expression.MockExpr{Expr: func(ctx session.Context, tuple btuple.Reader, schema bschema.Reader) expression.Value {
		// value in pos 0 is the most left column in index
		return bytes.Compare(tuple.ValueAt(0), []byte("bob")) == 0
	}}

	sujIndexScanPlan := plan.NewIndexScanPlan(model.NewIndexSchemaReader(mockDBInfo1.TableInfo[0], 0), indexPrefix, &mockExpr, 1, 1)

	// object index scan
	idxId = uint64(2)
	binary.BigEndian.PutUint64(indexId[:], idxId)
	// scan from
	objIndexPrefix := []byte(fmt.Sprintf("i%s_%s", indexId, "data2"))

	mockExpr2 := expression.MockExpr{Expr: func(ctx session.Context, tuple btuple.Reader, schema bschema.Reader) expression.Value {
		// value in pos 0 is the most left column in index
		return bytes.Compare(tuple.ValueAt(0), []byte("data2")) == 0
	}}

	objIndexScanPlan := plan.NewIndexScanPlan(model.NewIndexSchemaReader(mockDBInfo1.TableInfo[0], 1), objIndexPrefix, &mockExpr2, 1, 1)

	// multi scan plan
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

func BenchmarkMultiIndexScanExecutor(b *testing.B) {
	b.Run("basic", func(b *testing.B) {
		p := "./__test_tmp__/multi_index_scan_exec"
		mockDb := OpenMockDB(b, p)
		defer func() {
			mockDb.Close()
			os.RemoveAll(p)
		}()
		setupMockDB(b, mockDb)

		// insert tuples
		sc := mockDb.NewTxnAt(4, true)

		var pPolicies []value.Values
		for i := 0; i < 10_000_000; i++ {
			v := value.Values{
				value.NewStringValue(fmt.Sprintf("user%d", i)),
				value.NewStringValue(fmt.Sprintf("data%d", i)),
				value.NewStringValue("read"),
			}
			pPolicies = append(pPolicies, v)
		}

		_, _, err := mockDb.InsertTuples(b, sc, 1, 1, pPolicies)

		err = sc.CommitTxn(context.TODO(), 5)
		assert.Nil(b, err)
		assert.Nil(b, mockDb.WaitForMark(context.TODO(), 5))

		// multi-index scan
		// where subject = bob and object = data2
		sc = mockDb.NewTxnAt(6, true)
		builder := executorBuilder{ctx: sc}

		// subject index scan
		idxId := uint64(1)
		var indexId [8]byte
		binary.BigEndian.PutUint64(indexId[:], idxId)
		// scan from
		indexPrefix := []byte(fmt.Sprintf("i%s_%s", indexId, "user50001"))

		mockExpr := expression.MockExpr{Expr: func(ctx session.Context, tuple btuple.Reader, schema bschema.Reader) expression.Value {
			// value in pos 0 is the most left column in index
			return bytes.Compare(tuple.ValueAt(0), []byte("user50001")) == 0
		}}

		sujIndexScanPlan := plan.NewIndexScanPlan(model.NewIndexSchemaReader(mockDBInfo1.TableInfo[0], 0), indexPrefix, &mockExpr, 1, 1)

		// object index scan
		idxId = uint64(2)
		binary.BigEndian.PutUint64(indexId[:], idxId)
		// scan from
		objIndexPrefix := []byte(fmt.Sprintf("i%s_%s", indexId, "data999"))

		mockExpr2 := expression.MockExpr{Expr: func(ctx session.Context, tuple btuple.Reader, schema bschema.Reader) expression.Value {
			// value in pos 0 is the most left column in index
			return bytes.Compare(tuple.ValueAt(0), []byte("data999")) == 0
		}}

		objIndexScanPlan := plan.NewIndexScanPlan(model.NewIndexSchemaReader(mockDBInfo1.TableInfo[0], 1), objIndexPrefix, &mockExpr2, 1, 1)

		// multi scan plan
		multiIndexScanPlan := plan.NewMultiIndexScan([]plan.AbstractPlan{sujIndexScanPlan, objIndexScanPlan}, 1, 1)

		// execute
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				exec, _ := builder.Build(multiIndexScanPlan), builder.Error()
				Execute(exec, context.TODO())
			}
		})

	})

}
