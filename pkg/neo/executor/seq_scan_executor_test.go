package executor

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewSeqScanExecutor(t *testing.T) {
	p := "./__test_tmp__/seq_scan_exec"
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

	// seq scan tuples
	sc = mockDb.NewTxnAt(6, false)
	result, ids, err := mockDb.SeqScan(t, sc, 1, 1, mockDBInfo1.TableInfo[0])
	assert.Nil(t, sc.CommitTxn(context.TODO(), 7))
	assert.Nil(t, err)

	TuplesAsserter(t, inserted, result)
	IdsAsserter(t, insertedIds, ids)

}
