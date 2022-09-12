package model

import (
	"errors"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"strings"
)

type EffectPolicyType uint8

const (
	AllowOverride EffectPolicyType = iota
	DenyOverride
	AllowAndDeny
	Priority
	PriorityBaseOnRole
)

var (
	str2Type = []string{
		"some(where (p.eft == allow))",
		"!some(where (p.eft == deny))",
		"some(where (p.eft == allow)) && !some(where (p.eft == deny))",
		"priority(p.eft) || deny",
		"subjectPriority(p.eft)",
	}
	str2TypeMap               map[string]EffectPolicyType
	ErrInvalidEffectPolicyDef = errors.New("invalid effect policy definition")
)

func init() {
	str2TypeMap = map[string]EffectPolicyType{}
	for i, s2 := range str2Type {
		str2TypeMap[strings.ReplaceAll(s2, " ", "")] = EffectPolicyType(i)
	}
}

func NewEffectPolicyTypeFromString(s string) (EffectPolicyType, error) {
	if v, ok := str2TypeMap[strings.ReplaceAll(s, " ", "")]; ok {
		return v, nil
	}
	return 0, ErrInvalidEffectPolicyDef
}

type MatcherInfo struct {
	ID           uint64
	Name         CIStr
	Raw          string
	EffectPolicy EffectPolicyType
	Predicate    ast.Evaluable
}

func (m *MatcherInfo) Clone() *MatcherInfo {
	return &*m
}

func GenerateEffectPolicyAst(policyTable, eftColumnName, allow, deny string, policyType EffectPolicyType) []ast.Evaluable {
	Asts := [][]ast.Evaluable{
		{
			&ast.BinaryOperationExpr{
				Op: ast.EQ_OP,
				L: &ast.Accessor{
					Typ: ast.MEMBER_ACCESSOR, Ancestor: &ast.Primitive{Typ: ast.IDENTIFIER, Value: policyTable}, Ident: &ast.Primitive{Typ: ast.IDENTIFIER, Value: eftColumnName},
				},
				R: &ast.Primitive{Typ: ast.STRING, Value: allow},
			},
		},
		{
			&ast.BinaryOperationExpr{
				Op: ast.EQ_OP,
				L: &ast.Accessor{
					Typ: ast.MEMBER_ACCESSOR, Ancestor: &ast.Primitive{Typ: ast.IDENTIFIER, Value: policyTable}, Ident: &ast.Primitive{Typ: ast.IDENTIFIER, Value: eftColumnName},
				},
				R: &ast.Primitive{Typ: ast.STRING, Value: deny},
			},
		},
		{
			&ast.BinaryOperationExpr{
				Op: ast.EQ_OP,
				L: &ast.Accessor{
					Typ: ast.MEMBER_ACCESSOR, Ancestor: &ast.Primitive{Typ: ast.IDENTIFIER, Value: policyTable}, Ident: &ast.Primitive{Typ: ast.IDENTIFIER, Value: eftColumnName},
				},
				R: &ast.Primitive{Typ: ast.STRING, Value: allow},
			},
			&ast.BinaryOperationExpr{
				Op: ast.NE_OP,
				L: &ast.Accessor{
					Typ: ast.MEMBER_ACCESSOR, Ancestor: &ast.Primitive{Typ: ast.IDENTIFIER, Value: policyTable}, Ident: &ast.Primitive{Typ: ast.IDENTIFIER, Value: eftColumnName},
				},
				R: &ast.Primitive{Typ: ast.STRING, Value: deny},
			},
		},
		//TODO: priority
		//TODO: subjectPriority
	}
	return Asts[policyType]
}
