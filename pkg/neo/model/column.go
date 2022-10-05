package model

import "github.com/casbin-mesh/neo/pkg/primitive/bsontype"

type ColumnInfo struct {
	ID              uint64
	ColName         CIStr
	Offset          int
	Tp              bsontype.Type
	DefaultValue    Cloneable
	DefaultValueBit []byte
}

func (c *ColumnInfo) GetDefaultValue() []byte {
	return c.DefaultValueBit
}

func (c *ColumnInfo) Clone() *ColumnInfo {
	nc := *c
	if nc.DefaultValue != nil {
		nc.DefaultValue = c.DefaultValue.Clone()
	}
	if nc.DefaultValueBit != nil {
		nc.DefaultValueBit = append([]byte{}, nc.DefaultValueBit...)
	}
	return &nc
}

func (c *ColumnInfo) Name() []byte {
	return []byte(c.ColName.L)
}

func (c *ColumnInfo) Type() bsontype.Type {
	return c.Tp
}
