package expression

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/expression/iterator"
)

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
