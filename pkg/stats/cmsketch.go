// modified the https://github.com/pingcap/tidb/blob/master/statistics/cmsketch.go

package stats

import (
	"errors"
	"github.com/spaolacci/murmur3"
	"math"
	"sort"
)

// CMSketch is used to estimate point queries.
// Refer: https://en.wikipedia.org/wiki/Count-min_sketch
type CMSketch struct {
	depth int32
	width int32
	count uint64
	table [][]uint32
}

// NewCMSketch returns a new CM sketch.
func NewCMSketch(d, w int32) *CMSketch {
	tbl := make([][]uint32, d)
	arena := make([]uint32, d*w)

	for i := range tbl {
		tbl[i] = arena[i*int(w) : (i+1)*int(w)]
	}
	return &CMSketch{depth: d, width: w, table: tbl}
}

// InsertBytes inserts the bytes value into the CM Sketch.
func (c *CMSketch) InsertBytes(bytes []byte) {
	c.insertBytesByCount(bytes, 1)
}

// insertBytesByCount adds the bytes value into the TopN (if value already in TopN) or CM Sketch by delta, this does not updates c.defaultValue.
func (c *CMSketch) insertBytesByCount(bytes []byte, count uint64) {
	h1, h2 := murmur3.Sum128(bytes)
	c.count += count
	for i := range c.table {
		j := (h1 + h2*uint64(i)) % uint64(c.width)
		c.table[i][j] += uint32(count)
	}
}

// QueryBytes is used to query the count of specified bytes.
func (c *CMSketch) QueryBytes(d []byte) uint64 {
	h1, h2 := murmur3.Sum128(d)
	return c.queryHashValue(h1, h2)
}

func (c *CMSketch) queryHashValue(h1, h2 uint64) uint64 {
	vals := make([]uint32, c.depth)
	min := uint32(math.MaxUint32)
	// We want that when res is 0 before the noise is eliminated, the default value is not used.
	// So we need a temp value to distinguish before and after eliminating noise.
	temp := uint32(1)
	for i := range c.table {
		j := (h1 + h2*uint64(i)) % uint64(c.width)
		if min > c.table[i][j] {
			min = c.table[i][j]
		}
		noise := (c.count - uint64(c.table[i][j])) / (uint64(c.width) - 1)
		if uint64(c.table[i][j]) == 0 {
			vals[i] = 0
		} else if uint64(c.table[i][j]) < noise {
			vals[i] = temp
		} else {
			vals[i] = c.table[i][j] - uint32(noise) + temp
		}
	}
	sort.Slice(vals, func(i, j int) bool {
		return vals[i] < vals[j]
	})
	res := vals[(c.depth-1)/2] + (vals[c.depth/2]-vals[(c.depth-1)/2])/2
	if res > min+temp {
		res = min + temp
	}
	if res == 0 {
		return uint64(0)
	}
	res = res - temp
	return uint64(res)
}

// MergeCMSketch merges two CM Sketch.
func (c *CMSketch) MergeCMSketch(rc *CMSketch) error {
	if c == nil || rc == nil {
		return nil
	}
	if c.depth != rc.depth || c.width != rc.width {
		return errors.New("dimensions of Count-Min Sketch should be the same")
	}
	c.count += rc.count
	for i := range c.table {
		for j := range c.table[i] {
			c.table[i][j] += rc.table[i][j]
		}
	}
	return nil
}

// TotalCount returns the total count in the sketch, it is only used for test.
func (c *CMSketch) TotalCount() uint64 {
	return c.count
}

// Copy makes a copy for current CMSketch.
func (c *CMSketch) Copy() *CMSketch {
	if c == nil {
		return nil
	}
	tbl := make([][]uint32, c.depth)
	for i := range tbl {
		tbl[i] = make([]uint32, c.width)
		copy(tbl[i], c.table[i])
	}
	return &CMSketch{count: c.count, width: c.width, depth: c.depth, table: tbl}
}

// GetWidthAndDepth returns the width and depth of CM Sketch.
func (c *CMSketch) GetWidthAndDepth() (int32, int32) {
	return c.width, c.depth
}
