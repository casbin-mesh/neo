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
				return DBInfoKey(id)
			},
			expected: []byte(fmt.Sprintf("s_d%s", bid)),
		},
		{
			run: func() []byte {
				return IndexInfoKey(id)
			},
			expected: []byte(fmt.Sprintf("s_i%s", bid)),
		},
		{
			run: func() []byte {
				return MatcherInfoKey(id)
			},
			expected: []byte(fmt.Sprintf("s_m%s", bid)),
		},
		{
			run: func() []byte {
				return TableInfoKey(id)
			},
			expected: []byte(fmt.Sprintf("s_t%s", bid)),
		},
		{
			run: func() []byte {
				return MetaKey("test")
			},
			expected: []byte(fmt.Sprintf("m_ntest")),
		},
		{
			run: func() []byte {
				return TableKey(id, "test")
			},
			expected: []byte(fmt.Sprintf("m_d%s_ttest", bid)),
		},
		{
			run: func() []byte {
				return IndexKey(id, "test")
			},
			expected: []byte(fmt.Sprintf("m_t%s_itest", bid)),
		},
		{
			run: func() []byte {
				return MatcherKey(id, "test")
			},
			expected: []byte(fmt.Sprintf("m_d%s_mtest", bid)),
		},
	}
	runTestSets(t, sets)
}
