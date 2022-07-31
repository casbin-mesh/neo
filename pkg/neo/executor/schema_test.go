package executor

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSchemaExec_CreateDatabase(t *testing.T) {
	p := "./__test_tmp__/schema_exec"
	mockDb := OpenMockDB(t, p)
	defer func() {
		mockDb.Close()
		os.RemoveAll(p)
	}()
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
}
