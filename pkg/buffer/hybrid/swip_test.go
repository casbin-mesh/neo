package hybrid

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSwip(t *testing.T) {

	t.Run("should be hot", func(t *testing.T) {
		bf := BufferFrame{}
		pid := uint64(10)
		swip := NewSwip(pid, &bf)
		assert.Equal(t, uint64(10<<2), swip.pid)
		assert.True(t, swip.isHot())
	})
	t.Run("should be cool", func(t *testing.T) {
		bf := BufferFrame{}
		pid := uint64(10)
		swip := NewSwip(pid, &bf)
		swip.cool()
		assert.Equal(t, uint64((10<<2)|2), swip.pid)
		assert.True(t, swip.isCool())
	})
	t.Run("should be evicted", func(t *testing.T) {
		bf := BufferFrame{}
		pid := uint64(10)
		swip := NewSwip(pid, &bf)
		swip.evict(pid)
		assert.Equal(t, uint64((10<<2)|1), swip.pid)
		assert.True(t, swip.isEvicted())
	})
	t.Run("should be equal pid", func(t *testing.T) {
		bf := BufferFrame{}
		pid := uint64(10)
		swip := NewSwip(pid, &bf)
		assert.Equal(t, pid, swip.asPageId())
	})
}
