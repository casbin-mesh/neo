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

type PredicateType int

const (
	Or PredicateType = iota + 1
	And
	Other
)

type Predicate struct {
	Type PredicateType
	Args []Predicate
	Expr ast.Evaluable
}

func (p *Predicate) Clone() Predicate {
	np := *p
	np.Args = make([]Predicate, len(p.Args))
	for i, arg := range p.Args {
		np.Args[i] = arg.Clone()
	}
	np.Expr = p.Expr.Clone()
	return np
}

func NewPredicate(root ast.Evaluable) Predicate {
	switch root.(type) {
	case *ast.BinaryOperationExpr:
		node := root.(*ast.BinaryOperationExpr)
		if node.Op == ast.AND_OP {
			args := make([]Predicate, 0, 2)
			args = append(args, NewPredicate(node.L))
			args = append(args, NewPredicate(node.R))
			return Predicate{
				Type: And,
				Args: args,
			}
		} else if node.Op == ast.OR_OP {
			args := make([]Predicate, 0, 2)
			args = append(args, NewPredicate(node.L))
			args = append(args, NewPredicate(node.R))
			return Predicate{
				Type: Or,
				Args: args,
			}
		}
	}
	return Predicate{
		Type: Other,
		Expr: root,
	}
}

func RewritePredicate(predicate Predicate) Predicate {
	switch predicate.Type {
	case Or:
		args := make([]Predicate, 0, len(predicate.Args))
		for _, arg := range predicate.Args {
			rewritten := RewritePredicate(arg)
			args = append(args, rewritten)
		}
		flatten := FlatOrs(args)
		return Predicate{Type: Or, Args: flatten}
	case And:
		args := make([]Predicate, 0, len(predicate.Args))
		for _, arg := range predicate.Args {
			rewritten := RewritePredicate(arg)
			args = append(args, rewritten)
		}
		return Predicate{Type: And, Args: FlatAnds(args)}
	default:
		return predicate
	}
}

func FlatOrs(predicates []Predicate) (flatten []Predicate) {
	for _, pre := range predicates {
		switch pre.Type {
		case Or:
			flatten = append(flatten, FlatOrs(pre.Args)...)
		default:
			flatten = append(flatten, pre)
		}
	}
	return
}

func FlatAnds(predicates []Predicate) (flatten []Predicate) {
	for _, pre := range predicates {
		switch pre.Type {
		case And:
			flatten = append(flatten, FlatAnds(pre.Args)...)
		default:
			flatten = append(flatten, pre)
		}
	}
	return
}

// ConnectSubtree returns connected subtree
func ConnectSubtree(left, right ast.Evaluable) ast.Evaluable {
	return &ast.BinaryOperationExpr{Op: ast.AND_OP, L: left, R: right}
}

func AppendAst2Predicate(predicate *Predicate, ast ast.Evaluable, skip func(ast ast.Evaluable) bool) {
	switch predicate.Type {
	case And:
		predicate.Args = append(predicate.Args, NewPredicate(ast))
	case Or:
		for i, arg := range predicate.Args {
			switch arg.Type {
			case And:
				AppendAst2Predicate(&predicate.Args[i], ast, skip)
			case Or:
				AppendAst2Predicate(&predicate.Args[i], ast, skip)
			case Other:
				AppendAst2Predicate(&predicate.Args[i], ast, skip)
			}
		}
	case Other:
		if skip != nil && skip(predicate.Expr) {
			break
		}
		*predicate = Predicate{
			Type: And,
			Args: []Predicate{*predicate, NewPredicate(ast)},
		}
	}
}
