package planner

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/parser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGeneratePlanWithIndex(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		// multi-index

	})
}

type mockCtx struct {
	policyTableName         string
	effectColName           string
	allowIdent              string
	denyIdent               string
	reqAccessorAncestorName string
}

func (m mockCtx) PolicyTableName() string {
	return m.policyTableName
}

func (m mockCtx) EffectColName() string {
	return m.effectColName
}

func (m mockCtx) AllowIdent() string {
	return m.allowIdent
}

func (m mockCtx) DenyIdent() string {
	return m.denyIdent
}

func (m mockCtx) ReqAccessorAncestorName() string {
	return m.reqAccessorAncestorName
}

func NewMockCtx() Ctx {
	return &mockCtx{
		policyTableName:         "p",
		effectColName:           "eft",
		allowIdent:              "allow",
		denyIdent:               "deny",
		reqAccessorAncestorName: "r",
	}
}

type IsConstNodeSet struct {
	input    ast.Evaluable
	expected bool
}

func TestIsConstNode(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		ctx := NewMockCtx()
		sets := []IsConstNodeSet{
			{input: parser.MustParseFromString("1"), expected: true},
			{input: parser.MustParseFromString("\"root\""), expected: true},
			{input: parser.MustParseFromString("true"), expected: true},
			{input: parser.MustParseFromString("1.0"), expected: true},
			{input: &ast.Primitive{Typ: ast.NULL}, expected: true},
			{
				input: &ast.Accessor{
					Ancestor: &ast.Primitive{
						Typ:   ast.IDENTIFIER,
						Value: ctx.ReqAccessorAncestorName(),
					},
				},
				expected: true,
			},
		}
		for _, set := range sets {
			assert.Equal(t, set.expected, IsConstNode(ctx, set.input))
		}
	})
	t.Run("basic expr", func(t *testing.T) {
		ctx := NewMockCtx()
		sets := []IsConstNodeSet{
			{input: parser.MustParseFromString("r.sub==\"root\""), expected: true},
			{input: parser.MustParseFromString("r.sub==p.sub"), expected: false},
		}
		for _, set := range sets {
			assert.Equal(t, set.expected, IsConstNode(ctx, set.input))
		}
	})
}

var (
	allowOverride = model.GenerateEffectPolicyAst("p", "eft", "allow", "deny", model.AllowOverride)
	denyOverride  = model.GenerateEffectPolicyAst("p", "eft", "allow", "deny", model.DenyOverride)
)

type testMergePolicyEffect struct {
	ctx          Ctx
	predicate    ast.Evaluable
	policyEffect ast.Evaluable
	expectedOp   PredicateGroupType
	expected     []ast.Evaluable
}

func runTests(t *testing.T, sets []testMergePolicyEffect) {
	for _, set := range sets {
		pg := MergePolicyEffect(set.ctx, FlatExprTree(set.ctx, set.predicate), set.policyEffect)
		assert.Equal(t, set.expectedOp, pg.Op)
		assert.Equal(t, set.expected, pg.Cond)
	}
}

func getSets(policyEffect ast.Evaluable) []testMergePolicyEffect {
	ctx := NewMockCtx()
	// covers:
	// - basic_without_users_model.conf
	// - basic_without_resources_model.conf
	// - rbac_model.conf
	// - rbac_with_domains_model.conf
	// - abac_model.conf
	// - keymatch_model.conf
	// - rbac_with_not_deny_model.conf
	// - priority_model_explicit.conf
	// - subject_priority_model.conf
	var sets []testMergePolicyEffect
	setGenerator := func(predicates []string) {
		for _, predicate := range predicates {
			sets = append(sets, testMergePolicyEffect{
				ctx:          ctx,
				predicate:    parser.MustParseFromString(predicate),
				policyEffect: policyEffect,
				expectedOp:   AndPredicateGroup,
				expected: append(
					FlatExprTree(ctx, parser.MustParseFromString(predicate)).Cond,
					policyEffect,
				),
			})
		}
	}
	setGenerator([]string{
		"r.sub == p.sub && r.obj == p.obj && r.act == p.act",
		"r.obj == p.obj && r.act == p.act",
		"r.sub == p.sub && r.act == p.act",
		"g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act",
		"g(r.sub, p.sub) && g2(r.obj, p.obj) && r.act == p.act",
		"r.sub == r.obj.Owner",
		"r.sub == p.sub && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act)",
		"g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act",
	})
	// covers
	// - basic_with_root_model.conf
	sets = append(sets, testMergePolicyEffect{
		ctx:          ctx,
		predicate:    parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act || r.sub == \"root\""),
		policyEffect: policyEffect,
		expectedOp:   OrPredicateGroup,
		expected: []ast.Evaluable{
			parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act && p.eft==\"allow\""),
			parser.MustParseFromString("r.sub == \"root\""),
		},
	})

	return sets
}

func TestMergePolicyEffect(t *testing.T) {
	t.Run("basic test", func(t *testing.T) {
		sets := getSets(allowOverride[0])
		runTests(t, sets)
	})
	t.Run("basic test2", func(t *testing.T) {
		ctx := NewMockCtx()
		sets := append([]testMergePolicyEffect{}, testMergePolicyEffect{
			ctx:          ctx,
			predicate:    parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act || r.sub == \"root\""),
			policyEffect: allowOverride[0],
			expectedOp:   OrPredicateGroup,
			expected: []ast.Evaluable{
				parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act && p.eft==\"deny\""),
				parser.MustParseFromString("r.sub == \"root\""),
			},
		})
		sets = append([]testMergePolicyEffect{}, testMergePolicyEffect{
			ctx:          ctx,
			predicate:    parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act || r.sub == \"root\""),
			policyEffect: denyOverride[0],
			expectedOp:   OrPredicateGroup,
			expected: []ast.Evaluable{
				parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act && p.eft==\"deny\""),
				parser.MustParseFromString("r.sub == \"root\""),
			},
		})
		runTests(t, sets)
	})
}
