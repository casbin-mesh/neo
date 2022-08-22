package bschema

import (
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
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
	AppendFormField(f Field)
}

type Reader interface {
	FieldAt(pos int) Field
	FieldsLen() int
}

// readerWriter represents a bschema of a model, managing a table.
type readerWriter struct {
	fields []*field
	valLen int
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

func NewReaderWriteFormReader(r Reader) ReaderWriter {
	fields := make([]*field, r.FieldsLen())
	for i := 0; i < r.FieldsLen(); i++ {
		fields[i] = cloneField(r.FieldAt(i))
	}
	return &readerWriter{fields: fields}
}

func (bs *readerWriter) AppendFormField(f Field) {
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
