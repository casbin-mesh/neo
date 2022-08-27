package model

import "github.com/casbin-mesh/neo/pkg/primitive/bschema"

type IndexSchemaReader struct {
	table    *TableInfo
	index    *IndexInfo
	indexIdx int
}

func (i IndexSchemaReader) FieldAt(pos int) bschema.Field {
	indexCol := i.index.Columns[pos]
	return i.table.Columns[indexCol.Offset]
}

func (i IndexSchemaReader) FieldsLen() int {
	return len(i.index.Columns)
}

func NewIndexSchemaReader(info *TableInfo, indexIdx int) bschema.Reader {
	return &IndexSchemaReader{
		table:    info,
		index:    info.Indices[indexIdx],
		indexIdx: indexIdx,
	}
}
