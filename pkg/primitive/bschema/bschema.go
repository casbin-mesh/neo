package bschema

import (
	"bytes"
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
	"github.com/casbin-mesh/neo/pkg/utils/trick"
)

type BSchema interface {
	Key() []byte
	// ValueCopy returns a copy of the value of the item from the value bytes, writing it to dst slice.
	// If nil is passed, or capacity of dst isn't sufficient, a new slice would be allocated and
	// returned.
	ValueCopy(bytes []byte) ([]byte, error)
}

type ReaderWriter interface {
	Writer
	Reader
}

type Writer interface {
	Append(typ bsontype.Type, name []byte, defaultValue []byte)
	AppendFromField(f Field)
}

type Reader interface {
	Field(string) int
	FieldAt(pos int) Field
	FieldsLen() int
}

// readerWriter represents a bschema of a model, managing a table.
type readerWriter struct {
	fields []*field
	valLen int
}

func (bs *readerWriter) Field(s string) int {
	for i, f := range bs.fields {
		if bytes.Compare(trick.Slice(s), f.Name()) == 0 {
			return i
		}
	}
	return -1
}

func NewReaderWriter() ReaderWriter {
	return &readerWriter{}
}

func cloneField(f Field) *field {
	return &field{
		name:         f.Name(),
		typ:          f.Type(),
		defaultValue: f.GetDefaultValue(),
	}
}

func NewReaderWriteFromReader(r Reader) ReaderWriter {
	fields := make([]*field, r.FieldsLen())
	for i := 0; i < r.FieldsLen(); i++ {
		fields[i] = cloneField(r.FieldAt(i))
	}
	return &readerWriter{fields: fields}
}

func (bs *readerWriter) AppendFromField(f Field) {
	bs.fields = append(bs.fields, cloneField(f))
}

func (bs *readerWriter) Append(typ bsontype.Type, name []byte, defaultValue []byte) {
	bs.fields = append(bs.fields, &field{
		name:         name,
		typ:          typ,
		defaultValue: defaultValue,
	})
	bs.valLen += len(name) + 2 // 1B for type, 1B for NULL terminator
}

func (bs *readerWriter) FieldsLen() int { return len(bs.fields) }

func (bs *readerWriter) FieldAt(pos int) Field {
	return bs.fields[pos]
}
