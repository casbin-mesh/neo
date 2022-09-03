package expression

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/expression/iterator"
	"github.com/casbin-mesh/neo/pkg/neo/executor/expression"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
	"github.com/casbin-mesh/neo/pkg/primitive/value"
)

type Expression interface {
	Evaluate(ctx session.Context, tuple btuple.Reader, schema bschema.Reader) expression.Value
	AccessorMembers() []string
}

type AbstractExpression struct {
	base                  ast.Evaluable
	cachedAccessorMembers []string
}

func NewAbstractExpression(base ast.Evaluable) *AbstractExpression {
	return &AbstractExpression{
		base: base,
	}
}

func (a *AbstractExpression) initCachedAccessorMembers() {
	a.cachedAccessorMembers = GetAccessorMembers(a.base)
}

// AccessorMembers returns all accessor's members
func (a *AbstractExpression) AccessorMembers() []string {
	if a.cachedAccessorMembers == nil {
		a.initCachedAccessorMembers()
	}
	return a.cachedAccessorMembers
}

func (a *AbstractExpression) Evaluate(ctx session.Context, tuple btuple.Reader, schema bschema.Reader) expression.Value {
	return value.Value{}
}

func GetAccessorMembers(root ast.Evaluable) []string {
	iter := iterator.NewBfsIterator(root)
	var node ast.Evaluable
	nameSet := make(map[string]struct{})
	for {
		if node = iter.Next(); node == nil {
			break
		}
		if accessor, ok := node.(*ast.Accessor); ok {
			if ident, ok := accessor.Ident.(*ast.Primitive); ok {
				if name, ok := ident.Value.(string); ok {
					nameSet[name] = struct{}{}
				}
			}
		}
	}
	if len(nameSet) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(nameSet))
	for name, _ := range nameSet {
		result = append(result, name)
	}
	return result
}

// PruneSubtree returns Pruned node and remained node
func PruneSubtree(base ast.Evaluable, shouldBePruned func(evaluable ast.Evaluable) bool) (ast.Evaluable, ast.Evaluable) {
	iter := iterator.NewPostOrderIterator(&base, iterator.DefaultBinaryTreeTraversal)
	for {
		cur, parent, childIdx := iter.NextWithMutParent()
		// childIdx 0 or 1
		if cur == nil {
			break
		}
		if shouldBePruned(cur) {
			if parent != nil && *parent != nil {
				*parent = (*parent).GetChildAt(childIdx ^ 1)
				return cur, base
			}
			return cur, nil // pruned node is topmost node
		}
	}

	return nil, base
}

// ConnectSubtree returns connected subtree
func ConnectSubtree(left, right ast.Evaluable) ast.Evaluable {
	return &ast.BinaryOperationExpr{Op: ast.AND_OP, L: left, R: right}
}

func flatSubtree(root ast.Evaluable, filter func(node ast.Evaluable) bool) (result []ast.Evaluable) {
	iter := iterator.NewDfsIterator(root, &iterator.Option{Filter: filter})

	for {
		cur := iter.Next()
		if cur == nil {
			break
		}
		switch node := cur.(type) {
		case *ast.Primitive, *ast.Accessor, *ast.UnaryOperationExpr, *ast.ScalarFunction, *ast.TernaryOperationExpr:
			result = append(result, cur)
		case *ast.BinaryOperationExpr:
			// find a child node that not be filtered
			if node.L != nil && !filter(node.L) {
				result = append(result, node.L)
			}
			if node.R != nil && !filter(node.R) {
				result = append(result, node.R)
			}
		}
	}
	return
}

func FlatAndSubtree(root ast.Evaluable) []ast.Evaluable {
	return flatSubtree(root, func(node ast.Evaluable) bool {
		n, ok := node.(*ast.BinaryOperationExpr)
		return ok && n.Op == ast.AND_OP
	})
}

func FlatOrSubtree(root ast.Evaluable) []ast.Evaluable {
	return flatSubtree(root, func(node ast.Evaluable) bool {
		n, ok := node.(*ast.BinaryOperationExpr)
		return ok && n.Op == ast.OR_OP
	})
}
