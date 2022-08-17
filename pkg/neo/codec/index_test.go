package codec

import (
	"encoding/binary"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrimaryIndexEntryKey(t *testing.T) {
	id := uint64(1)
	bid := [8]byte{}
	binary.BigEndian.PutUint64(bid[:], id)
	assert.Equal(t, []byte(fmt.Sprintf("i%s_hello", bid)), PrimaryIndexEntryKey(1, []byte("hello")))
}

func TestSecondaryIndexEntryKey(t *testing.T) {
	id := uint64(1)
	bid := [8]byte{}
	binary.BigEndian.PutUint64(bid[:], id)
	assert.Equal(t, []byte(fmt.Sprintf("i%s_hello_%s", bid, bid)), SecondaryIndexEntryKey(1, []byte("hello"), bid[:]))
}
