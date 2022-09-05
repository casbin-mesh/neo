package model

import (
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

func (t *TableInfo) FieldAt(pos int) bschema.Field {
	return t.Columns[pos]
}

func (t *TableInfo) FieldsLen() int {
	return len(t.Columns)
}
