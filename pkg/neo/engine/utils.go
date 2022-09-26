// Copyright 2022 The casbin-neo Authors. All Rights Reserved.
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

package engine

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/codec"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
	"github.com/casbin-mesh/neo/pkg/primitive/value"
	"golang.org/x/exp/slices"
	"io"
	"sync/atomic"
	"time"
)

var sessionIdCounter = readRandomUint32()
var processUnique = processUniqueBytes()

func processUniqueBytes() [5]byte {
	var b [5]byte
	_, err := io.ReadFull(rand.Reader, b[:])
	if err != nil {
		panic(fmt.Errorf("cannot initialize objectid package with crypto.rand.Reader: %v", err))
	}

	return b
}

func readRandomUint32() uint32 {
	var b [4]byte
	_, err := io.ReadFull(rand.Reader, b[:])
	if err != nil {
		panic(fmt.Errorf("cannot initialize objectid package with crypto.rand.Reader: %v", err))
	}

	return (uint32(b[0]) << 0) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16) | (uint32(b[3]) << 24)
}

func NewSessionId() string {
	var b [12]byte

	binary.BigEndian.PutUint32(b[0:4], uint32(time.Now().Unix()))
	copy(b[4:9], processUnique[:])
	putUint24(b[9:12], atomic.AddUint32(&sessionIdCounter, 1))

	return Hex(b)
}

func Hex(id [12]byte) string {
	var buf [24]byte
	hex.Encode(buf[:], id[:])
	return string(buf[:])
}

func putUint24(b []byte, v uint32) {
	b[0] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[2] = byte(v)
}

func A2Values(a A) value.Values {
	result := make([]value.Value, 0, len(a))
	for _, i2 := range a {
		result = append(result, value.NewValueFromInterface(i2))
	}
	return result
}

func A2ValuesArray(a []A) []value.Values {
	result := make([]value.Values, 0, len(a))
	for _, a2 := range a {
		result = append(result, A2Values(a2))
	}
	return result
}

var (
	ErrInvalidValueLen = errors.New("invalid value len")
)

func DecodeValue2Map(modifier btuple.Modifier, schema bschema.Reader) (M, error) {
	if len(modifier.Values()) != schema.FieldsLen() {
		return nil, ErrInvalidValueLen
	}
	result := M{}
	for i, elem := range modifier.Values() {
		key := string(schema.FieldAt(i).Name())
		result[key] = codec.DecodeValue2NaiveType(elem, schema.FieldAt(i).Type())
	}
	return result, nil
}

func DecodeValues2Map(modifier []btuple.Modifier, schema bschema.Reader) ([]M, error) {
	result := make([]M, 0, len(modifier))
	for _, m := range modifier {
		r, err := DecodeValue2Map(m, schema)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func DecodeValue(modifier btuple.Modifier, schema bschema.Reader) (A, error) {
	if len(modifier.Values()) != schema.FieldsLen() {
		return nil, ErrInvalidValueLen
	}
	result := make([]interface{}, 0, len(modifier.Values()))
	for i, elem := range modifier.Values() {
		result = append(result, codec.DecodeValue2NaiveType(elem, schema.FieldAt(i).Type()))
	}
	return result, nil
}

func DecodeValues(modifier []btuple.Modifier, schema bschema.Reader) ([]A, error) {
	result := make([]A, 0, len(modifier))
	for _, m := range modifier {
		r, err := DecodeValue(m, schema)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

type ValueAccessor struct {
	table  *model.TableInfo
	lookup map[string]*ast.Primitive
}

func Bsontype2Asttype(p bsontype.Type) ast.Type {
	switch p {
	case bsontype.String:
		return ast.STRING
	default:
		return ast.ERROR
	}
}

func NewValueAccessor(table *model.TableInfo, data A) *ValueAccessor {
	lookup := map[string]*ast.Primitive{}
	for _, column := range table.Columns {
		lookup[column.ColName.L] = &ast.Primitive{
			Typ:   Bsontype2Asttype(column.Tp),
			Value: data[column.Offset],
		}
	}
	return &ValueAccessor{table: table, lookup: lookup}
}

func NewValueAccessorFromMap(table *model.TableInfo, filter map[string]interface{}) (*ValueAccessor, error) {
	lookup := map[string]*ast.Primitive{}
	for col, val := range filter {
		if idx := slices.IndexFunc(table.Columns, func(c *model.ColumnInfo) bool { return c.ColName.L == col }); idx != -1 {
			column := table.Columns[idx]
			err := checkType(column.Tp, val)
			if err != nil {
				return nil, err
			}
			lookup[column.ColName.L] = &ast.Primitive{
				Typ:   Bsontype2Asttype(column.Tp),
				Value: val,
			}
		}
	}
	return &ValueAccessor{table: table, lookup: lookup}, nil
}

func (v ValueAccessor) GetMember(ident string) *ast.Primitive {
	p, ok := v.lookup[ident]
	if ok {
		return p
	}
	return &ast.Primitive{Typ: ast.NULL}
}

var (
	ErrInvalidDataLen  = errors.New("invalid data length")
	ErrUnsupportedType = errors.New("unsupported type")
)

func checkType(t bsontype.Type, data interface{}) error {
	switch t {
	case bsontype.String:
		_, ok := data.(string)
		if !ok {
			return fmt.Errorf("expected type: %s, but got: %v", bsontype.String, data)
		}
		return nil
	default:
		return ErrUnsupportedType
	}
}
func preCheck(table *model.TableInfo, data A) error {
	if len(table.Columns) != len(data) {
		return ErrInvalidDataLen
	}
	for i, column := range table.Columns {
		err := checkType(column.Tp, data[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func Value2Accessor(table *model.TableInfo, data A) (ast.AccessorValue, error) {
	if err := preCheck(table, data); err != nil {
		return nil, err
	}
	return NewValueAccessor(table, data), nil
}
