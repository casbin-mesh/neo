package model

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
	"strings"
)

type TableInfo struct {
	ID          uint64
	Name        CIStr
	Columns     []*ColumnInfo
	Indices     []*IndexInfo
	ForeignKeys []*FKInfo
}

func (t *TableInfo) Field(s string) int {
	for i, col := range t.Columns {
		if strings.Compare(col.ColName.L, s) == 0 {
			return i
		}
	}
	return -1
}

func (t *TableInfo) Clone() *TableInfo {
	nt := *t
	nt.Columns = make([]*ColumnInfo, len(t.Columns))
	nt.Indices = make([]*IndexInfo, len(t.Indices))
	nt.ForeignKeys = make([]*FKInfo, len(t.ForeignKeys))
	for i, column := range t.Columns {
		nt.Columns[i] = column.Clone()
	}
	for i, index := range t.Indices {
		nt.Indices[i] = index.Clone()
	}
	for i, key := range t.ForeignKeys {
		nt.ForeignKeys[i] = key.Clone()
	}
	return &nt
}

func (t *TableInfo) SelectAst(reqName string) ast.Evaluable {
	tableName := t.Name.L
	tableAccessorAncestor := &ast.Primitive{Typ: ast.IDENTIFIER, Value: tableName}
	reqAccessorAncestor := &ast.Primitive{Typ: ast.IDENTIFIER, Value: reqName}
	var cur ast.Evaluable

	for _, column := range t.Columns {
		attrName := column.ColName.L
		attrIdent := &ast.Primitive{Typ: ast.IDENTIFIER, Value: attrName}
		node := &ast.BinaryOperationExpr{
			Op: ast.EQ_OP,
			L:  &ast.Accessor{Typ: ast.MEMBER_ACCESSOR, Ancestor: reqAccessorAncestor, Ident: attrIdent},
			R:  &ast.Accessor{Typ: ast.MEMBER_ACCESSOR, Ancestor: tableAccessorAncestor, Ident: attrIdent},
		}
		if cur == nil {
			cur = node
		} else {
			cur = &ast.BinaryOperationExpr{
				Op: ast.AND_OP,
				L:  cur,
				R:  node,
			}
		}
	}

	return cur
}

func (t *TableInfo) FieldAt(pos int) bschema.Field {
	return t.Columns[pos]
}

func (t *TableInfo) FieldsLen() int {
	return len(t.Columns)
}
