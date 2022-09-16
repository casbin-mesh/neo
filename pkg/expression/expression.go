package expression

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/expression/iterator"
	"github.com/casbin-mesh/neo/pkg/neo/codec"
	"github.com/casbin-mesh/neo/pkg/neo/executor/expression"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type Expression interface {
	Evaluate(ctx session.Context, evalCtx ast.EvaluateCtx, tuple btuple.Reader, schema bschema.Reader) (expression.Value, error)
	AccessorMembers() []string
}

type AbstractExpression struct {
	base                  ast.Evaluable
	cachedAccessorMembers []string
}

type TupleAccessor struct {
	tuple  btuple.Reader
	schema bschema.Reader
}

func (t TupleAccessor) GetMember(ident string) *ast.Primitive {
	if idx := t.schema.Field(ident); idx >= 0 {
		value := codec.DecodeValue(t.tuple.ValueAt(idx), t.schema.FieldAt(idx).Type())
		switch value.Type() {
		case bsontype.String:
			return &ast.Primitive{
				Typ:   ast.STRING,
				Value: value.GetString(),
			}
			// TODO: to support more types
		}
	}
	return &ast.Primitive{Typ: ast.NULL}
}

type MemoExpression struct {
	base     *AbstractExpression
	accessor *TupleAccessor
}

func (m *MemoExpression) AccessorMembers() []string {
	return m.base.AccessorMembers()
}

func NewExpression(base ast.Evaluable) (Expression, *TupleAccessor) {
	accessor := &TupleAccessor{}
	return &MemoExpression{
		base:     NewAbstractExpression(base),
		accessor: accessor,
	}, accessor
}

func (m *MemoExpression) Evaluate(ctx session.Context, evalCtx ast.EvaluateCtx, tuple btuple.Reader, schema bschema.Reader) (expression.Value, error) {
	m.accessor.schema = schema
	m.accessor.tuple = tuple
	return m.base.Evaluate(ctx, evalCtx, tuple, schema)
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

func (a *AbstractExpression) Evaluate(ctx session.Context, evalCtx ast.EvaluateCtx, tuple btuple.Reader, schema bschema.Reader) (expression.Value, error) {
	value, err := a.base.Evaluate(evalCtx)
	if err != nil {
		return nil, err
	}
	return value, nil
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
