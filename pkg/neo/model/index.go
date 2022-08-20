package model

type IndexType uint8

type IndexColumn struct {
	ColName CIStr
	Offset  int
}

func (i *IndexColumn) Clone() *IndexColumn {
	ni := *i
	return &ni
}

type IndexInfo struct {
	ID      uint64
	Name    CIStr
	Table   CIStr
	Columns []*IndexColumn
	Unique  bool
	Primary bool
	Tp      IndexType
}

func (i *IndexInfo) Leftmost() *IndexColumn {
	return i.Columns[0]
}

func (i *IndexInfo) Clone() *IndexInfo {
	ni := *i
	ni.Columns = make([]*IndexColumn, len(i.Columns))
	for j, column := range i.Columns {
		ni.Columns[j] = column.Clone()
	}
	return &ni
}
