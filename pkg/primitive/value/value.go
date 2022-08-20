package value

import (
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
	"github.com/casbin-mesh/neo/pkg/utils/trick"
)

type Values []Value

func (v *Values) Clone() Values {
	ret := make(Values, len(*v))
	for i, value := range ret {
		ret[i] = *value.Clone()
	}
	return ret
}

type Value struct {
	t         bsontype.Type
	collation uint8  // uint8
	length    uint32 // uint32
	i         int64  // int64 uint64 float64
	b         []byte // holds string or bytes
}

func (v *Value) Clone() *Value {
	ret := *v
	if v.b != nil {
		ret.b = make([]byte, len(v.b))
		copy(ret.b, v.b)
	}
	return &ret
}

func (v *Value) Type() bsontype.Type {
	return v.t
}

func (v *Value) GetBytes() []byte {
	return v.b
}

func (v *Value) String() string {
	return v.t.String()
}

// sink prevents s from being allocated on the stack.
var sink = func(s string) {
}

func NewStringValue(s string) Value {
	sink(s)
	return Value{
		t: bsontype.String,
		b: trick.Slice(s),
	}
}

func (v *Value) GetString() string {
	return string(trick.String(v.b))
}
