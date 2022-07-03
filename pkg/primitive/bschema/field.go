package bschema

import "github.com/casbin-mesh/neo/pkg/primitive/bsontype"

type Field struct {
	name []byte
	typ  bsontype.Type
}

func (f *Field) Type() bsontype.Type {
	return f.typ
}

func (f *Field) Name() []byte {
	return f.name
}

// Encode into binary format.
//
// | typ bsontype.Type |  name []byte |
func (f *Field) Encode() []byte {
	dst := make([]byte, f.len())
	dst[0] = byte(f.typ)
	copy(dst[1:], f.name)
	return dst
}

func (f *Field) len() int {
	return len(f.name) + 1 // 1 byte for type
}

// Decode from bytes
func (f *Field) Decode(src []byte) {
	f.typ = bsontype.Type(src[0])
	//TODO(weny): should we clone the src here?
	f.name = make([]byte, len(src)-1)
	copy(f.name[:], src[1:])
}
