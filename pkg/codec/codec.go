package codec

const ()

var (
	Separator = []byte{byte(0)}
)

type encoder interface {
	encodeKey() []byte
	encodeVal() []byte
}

type decoder interface {
	decode([]byte) interface{}
}
