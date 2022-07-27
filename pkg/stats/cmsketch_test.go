package stats

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func encodeInt64(val int64) []byte {
	var bytes [8]byte
	binary.BigEndian.PutUint64(bytes[:], uint64(val))

	return bytes[:]
}

func (c *CMSketch) insert(val int64) error {
	c.InsertBytes(encodeInt64(val))
	return nil
}

// buildCMSketchAndMapWithOffset builds cm sketch using zipf and the generated values starts from `offset`.
func buildCMSketchAndMapWithOffset(d, w int32, seed int64, total, imax uint64, s float64, offset int64) (*CMSketch, map[int64]uint32, error) {
	cms := NewCMSketch(d, w)
	mp := make(map[int64]uint32)
	zipf := rand.NewZipf(rand.New(rand.NewSource(seed)), s, 1, imax)
	for i := uint64(0); i < total; i++ {
		val := int64(zipf.Uint64()) + offset
		err := cms.insert(val)
		if err != nil {
			return nil, nil, err
		}
		mp[val]++
	}
	return cms, mp, nil
}

func buildCMSketchAndMap(d, w int32, seed int64, total, imax uint64, s float64) (*CMSketch, map[int64]uint32, error) {
	return buildCMSketchAndMapWithOffset(d, w, seed, total, imax, s, 0)
}

func averageAbsoluteError(cms *CMSketch, mp map[int64]uint32) (uint64, error) {
	var total uint64
	for num, count := range mp {
		estimate := cms.QueryBytes(encodeInt64(num))
		var diff uint64
		if uint64(count) > estimate {
			diff = uint64(count) - estimate
		} else {
			diff = estimate - uint64(count)
		}
		total += diff
	}
	return total / uint64(len(mp)), nil
}

func TestCMSketch(t *testing.T) {
	tests := []struct {
		zipfFactor float64
		avgError   uint64
	}{
		{
			zipfFactor: 1.1,
			avgError:   3,
		},
		{
			zipfFactor: 2,
			avgError:   24,
		},
		{
			zipfFactor: 3,
			avgError:   63,
		},
	}
	d, w := int32(5), int32(2048)
	total, imax := uint64(100000), uint64(1000000)
	for _, tt := range tests {
		lSketch, lMap, err := buildCMSketchAndMap(d, w, 0, total, imax, tt.zipfFactor)
		assert.NoError(t, err)
		avg, err := averageAbsoluteError(lSketch, lMap)
		assert.NoError(t, err)
		assert.LessOrEqual(t, avg, tt.avgError)

		rSketch, rMap, err := buildCMSketchAndMap(d, w, 1, total, imax, tt.zipfFactor)
		assert.NoError(t, err)
		avg, err = averageAbsoluteError(rSketch, rMap)
		assert.NoError(t, err)
		assert.LessOrEqual(t, avg, tt.avgError)

		err = lSketch.MergeCMSketch(rSketch)
		assert.NoError(t, err)
		for val, count := range rMap {
			lMap[val] += count
		}
		avg, err = averageAbsoluteError(lSketch, lMap)
		assert.NoError(t, err)
		assert.Less(t, avg, tt.avgError*2)
	}
}
