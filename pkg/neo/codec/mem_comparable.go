package codec

import (
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
	"github.com/casbin-mesh/neo/pkg/primitive/value"
)

func EncodeCmpValue(v value.Value) []byte {
	switch v.Type() {
	case bsontype.String, bsontype.Binary:
		return v.GetBytes()
		//TODO: to support more types
	}
	return nil
}
