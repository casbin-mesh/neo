package codec

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
	"github.com/casbin-mesh/neo/pkg/primitive/value"
)

func EncodePrimitive(v *ast.Primitive) []byte {
	switch v.Typ {
	case ast.STRING:
		return []byte(v.Value.(string))
		// TODO: add more types
	}
	return nil
}

func EncodeValue(v value.Value) []byte {
	switch v.Type() {
	case bsontype.String, bsontype.Binary:
		return v.GetBytes()
		//TODO: to support more types
	}
	return nil
}

func EncodeValues(vs value.Values) [][]byte {
	ret := make([][]byte, len(vs))
	for i, v := range vs {
		ret[i] = EncodeValue(v)
	}
	return ret
}

func DecodeValue2NaiveType(bytes []byte, p bsontype.Type) interface{} {
	switch p {
	case bsontype.String:
		return string(bytes)
		//TODO: to support more types
	}
	return value.Value{}
}

func DecodeValue(bytes []byte, p bsontype.Type) value.Value {
	switch p {
	case bsontype.String:
		v := value.NewStringValue(string(bytes))
		return v
		//TODO: to support more types
	}
	return value.Value{}
}
