// Copyright 2022 The casbin-mesh Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
)

type Cloneable interface {
	Clone() Cloneable
}

// CIStr is case insensitive string.
type CIStr struct {
	O string
	L string
}

type ColumnInfo struct {
	ID              uint64
	ColName         CIStr
	Tp              bsontype.Type
	DefaultValue    Cloneable
	DefaultValueBit []byte
}

func (c *ColumnInfo) Name() []byte {
	return []byte(c.ColName.L)
}

func (c *ColumnInfo) Type() bsontype.Type {
	return c.Tp
}

type IndexType uint8

type IndexInfo struct {
	ID      uint64
	Name    CIStr
	Table   CIStr
	Unique  bool
	Primary bool
	Tp      IndexType
}

type FKInfo struct {
	ID       uint64
	Name     CIStr
	RefTable CIStr
	RefCols  []CIStr
	Cols     []CIStr
	OnDelete int64
	OnUpdate int64
}

type TableInfo struct {
	ID          uint64
	Name        CIStr
	Columns     []*ColumnInfo
	Indices     []*IndexInfo
	ForeignKeys []*FKInfo
}

type MatcherInfo struct {
	ID           uint64
	Name         CIStr
	Raw          string
	EffectPolicy byte
}

type DBInfo struct {
	ID          uint64
	Name        CIStr
	TableInfo   []*TableInfo
	MatcherInfo []*MatcherInfo
}

func (f *FKInfo) Clone() *FKInfo {
	nf := *f
	nf.RefCols = make([]CIStr, len(f.RefCols))
	nf.Cols = make([]CIStr, len(f.Cols))
	copy(nf.RefCols, f.RefCols)
	copy(nf.Cols, f.Cols)
	return &nf
}

func (i *IndexInfo) Clone() *IndexInfo {
	return &*i
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

func (c *ColumnInfo) Clone() *ColumnInfo {
	nc := *c
	if nc.DefaultValue != nil {
		nc.DefaultValue = c.DefaultValue.Clone()
	}
	nc.DefaultValueBit = append([]byte{}, nc.DefaultValueBit...)
	return &nc
}

func (m *MatcherInfo) Clone() *MatcherInfo {
	return &*m
}

func (d *DBInfo) Clone() *DBInfo {
	nd := *d
	nd.MatcherInfo = make([]*MatcherInfo, len(d.MatcherInfo))
	for i, info := range d.MatcherInfo {
		nd.MatcherInfo[i] = info.Clone()
	}
	nd.TableInfo = make([]*TableInfo, len(d.TableInfo))
	for i, info := range d.TableInfo {
		nd.TableInfo[i] = info.Clone()
	}
	return &nd
}
