package model

import "github.com/casbin-mesh/neo/pkg/expression/ast"

type EffectPolicyType uint8

const (
	AllowOverride EffectPolicyType = iota
	DenyOverride
	AllowAndDeny
	Priority
	PriorityBaseOnRole
)

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
				R: &ast.Primitive{Typ: ast.STRING, Value: deny},
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
