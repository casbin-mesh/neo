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

// CIStr is case insensitive string.
type CIStr struct {
	O string
	L string
}

type ColumnInfo struct {
	ID              uint64
	Name            CIStr
	DefaultValue    interface{}
	DefaultValueBit []byte
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
