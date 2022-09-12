package planner

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
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
