package iterator

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDfsIterator_Next(t *testing.T) {
	tree := &ast.BinaryOperationExpr{
		Op: ast.AND_OP,
		L: &ast.BinaryOperationExpr{
			Op: ast.AND_OP,
			L: &ast.Primitive{
				Typ:   ast.INT,
				Value: 1,
			},
			R: &ast.Primitive{
				Typ:   ast.INT,
				Value: 2,
			},
		},
		R: &ast.BinaryOperationExpr{
			Op: ast.AND_OP,
			L: &ast.Primitive{
				Typ:   ast.INT,
				Value: 3,
			},
			R: &ast.Primitive{
				Typ:   ast.INT,
				Value: 4,
			},
		},
	}

	iter := NewDfsIterator(tree)

	// root
	n := iter.Next()
	expr, ok := n.(*ast.BinaryOperationExpr)
	assert.True(t, ok)
	assert.Equal(t, expr.Op, ast.AND_OP)

	// left and
	n = iter.Next()
	expr, ok = n.(*ast.BinaryOperationExpr)
	assert.True(t, ok)
	assert.Equal(t, expr.Op, ast.AND_OP)

	// left 1
	n = iter.Next()
	pri, ok := n.(*ast.Primitive)
	assert.True(t, ok)
	assert.Equal(t, pri.Value, 1)

	// right 2
	n = iter.Next()
	pri, ok = n.(*ast.Primitive)
	assert.True(t, ok)
	assert.Equal(t, pri.Value, 2)

	// right and
	n = iter.Next()
	expr, ok = n.(*ast.BinaryOperationExpr)
	assert.True(t, ok)
	assert.Equal(t, expr.Op, ast.AND_OP)

	// left 3
	n = iter.Next()
	pri, ok = n.(*ast.Primitive)
	assert.True(t, ok)
	assert.Equal(t, pri.Value, 3)

	// right 4
	n = iter.Next()
	pri, ok = n.(*ast.Primitive)
	assert.True(t, ok)
	assert.Equal(t, pri.Value, 4)

	assert.False(t, iter.HasNext())
}
