package schema

import "github.com/casbin-mesh/neo/pkg/codec"

type SchemaType byte

const (
	SchemaTypePolicy = 0
	SchemaTypeGroup  = 1
)

// BSchema represents a schema of a model, managing a table.
type BSchema struct {
	typ       SchemaType // policy or group
	name      []byte
	namespace []byte
	fields    []Field
}

func NewBSchema(namespace, name []byte) *BSchema {
	return &BSchema{namespace: namespace, name: name}
}

func (bs *BSchema) Append(typ FieldType, name []byte) {
	bs.fields = append(bs.fields, Field{
		name: name,
		typ:  typ,
	})
}

func (bs *BSchema) Namespace() []byte { return bs.namespace }

func (bs *BSchema) FieldsLen() int { return len(bs.fields) }

func (bs *BSchema) FieldAt(pos int) Field {
	return bs.fields[pos]
}

func (bs *BSchema) encodeKey() []byte {
	buf := make([]byte, 0, len(bs.namespace)+1+len(bs.name))
	buf = appendBytes(buf, bs.namespace, codec.Separator, bs.name)
	return buf
}

func (bs *BSchema) encodeVal() []byte {
	buf := make([]byte, 0) // TODO: optimize later
	buf = appendBytes(buf, bs.namespace, codec.Separator, bs.name, codec.Separator)
	for _, f := range bs.fields {
		buf = appendBytes(buf, []byte{byte(f.typ)}, f.name, codec.Separator)
	}
	return buf
}

func appendBytes(buf []byte, elem ...[]byte) []byte {
	for _, b := range elem {
		buf = append(buf, b...)
	}
	return buf
}

func (bs *BSchema) decode([]byte) {}
