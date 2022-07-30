package model

type IndexType uint8

type IndexInfo struct {
	ID      uint64
	Name    CIStr
	Table   CIStr
	Unique  bool
	Primary bool
	Tp      IndexType
}

func (i *IndexInfo) Clone() *IndexInfo {
	return &*i
}
