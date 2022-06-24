package schema

type FieldType byte

const (
	FieldTypeBinary = 1
)

type Field struct {
	name []byte
	typ  FieldType
}
