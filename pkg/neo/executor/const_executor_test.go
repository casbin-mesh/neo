package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/expression"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/parser"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type mockAccessorValue struct {
}

func (m *mockAccessorValue) GetMember(ident string) *ast.Primitive {
	switch ident {
	case "sub":
		return &ast.Primitive{Typ: ast.STRING, Value: "root"}
	default:
		return &ast.Primitive{Typ: ast.INT, Value: 0}
	}
}

func TestConstExecutor_Next(t *testing.T) {
	p := "./__test_tmp__/const_exec"
	mockDb := OpenMockDB(t, p)
	defer func() {
		mockDb.Close()
		os.RemoveAll(p)
	}()
	setupMockDB(t, mockDb)
	sc := mockDb.NewTxnAt(1, false)
	ctx := ast.NewContext()
	ctx.AddAccessor("r", &mockAccessorValue{})

	builder := executorBuilder{ctx: sc}
	expr := expression.NewAbstractExpression(parser.MustParseFromString("r.sub == \"root\""))
	exec := builder.Build(plan.NewConstPlan(expr, ctx))
	result, _, err := Execute(exec, context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, btuple.Elem{1}, result[0].ValueAt(0))
}
