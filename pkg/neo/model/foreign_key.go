package model

type FKInfo struct {
	ID       uint64
	Name     CIStr
	RefTable CIStr
	RefCols  []CIStr
	Cols     []CIStr
	OnDelete int64
	OnUpdate int64
}

func (f *FKInfo) Clone() *FKInfo {
	nf := *f
	nf.RefCols = make([]CIStr, len(f.RefCols))
	nf.Cols = make([]CIStr, len(f.Cols))
	copy(nf.RefCols, f.RefCols)
	copy(nf.Cols, f.Cols)
	return &nf
}
