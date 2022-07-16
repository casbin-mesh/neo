package codec

import (
	"encoding/binary"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestSet struct {
	run      func() []byte
	expected []byte
}

func runTestSets(t *testing.T, set []TestSet) {
	for _, testSet := range set {
		obtained := testSet.run()
		assert.Equal(t, len(testSet.expected), len(obtained))
		assert.Equal(t, testSet.expected, obtained)
	}
}

func TestKeyEncode(t *testing.T) {
	id := uint64(1)
	bid := [8]byte{}
	binary.BigEndian.PutUint64(bid[:], id)
	sets := []TestSet{
		{
			run: func() []byte {
				return MetaKey("test")
			},
			expected: []byte(fmt.Sprintf("m_ntest")),
		},
		{
			run: func() []byte {
				return TableKey("test")
			},
			expected: []byte(fmt.Sprintf("m_ttest")),
		},
		{
			run: func() []byte {
				return IndexKey("test")
			},
			expected: []byte(fmt.Sprintf("m_itest")),
		},
		{
			run: func() []byte {
				return MatcherKey("test")
			},
			expected: []byte(fmt.Sprintf("m_mtest")),
		},
	}
	runTestSets(t, sets)
}
