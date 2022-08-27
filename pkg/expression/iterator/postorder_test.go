package iterator

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPostorderIterator_Next(t *testing.T) {
	var tree ast.Evaluable
	tree = &ast.BinaryOperationExpr{
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

	iter := NewPostOrderIterator(&tree, &Option{func(node ast.Evaluable) bool {
		_, ok := node.(*ast.BinaryOperationExpr)
		return ok
	}})

	// leftmost subtree
	cur, parent, idx := iter.NextWithMutParent()
	assert.Equal(t, &ast.BinaryOperationExpr{
		Op: ast.AND_OP,
		L: &ast.Primitive{
			Typ:   ast.INT,
			Value: 1,
		},
		R: &ast.Primitive{
			Typ:   ast.INT,
			Value: 2,
		},
	}, cur)
	assert.Equal(t, tree, *parent)
	assert.Equal(t, 0, idx)

	// rightmost subtree
	cur, parent, idx = iter.NextWithMutParent()
	assert.Equal(t, &ast.BinaryOperationExpr{
		Op: ast.AND_OP,
		L: &ast.Primitive{
			Typ:   ast.INT,
			Value: 3,
		},
		R: &ast.Primitive{
			Typ:   ast.INT,
			Value: 4,
		},
	}, cur)
	assert.Equal(t, tree, *parent)
	assert.Equal(t, 1, idx)
	// top
	cur, parent, idx = iter.NextWithMutParent()
	assert.Equal(t, tree, cur)
	assert.Nil(t, parent)

	// last
	assert.False(t, iter.HasNext())
}
