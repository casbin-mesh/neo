package simple

import (
	"github.com/casbin-mesh/neo/pkg/storage/disk"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"unsafe"
)

func TestDefault(t *testing.T) {
	m := Default()
	assert.Nilf(t, m.Open("test.db"), "Failed to open file")
	data := [disk.PAGE_SIZE]byte{1, 2, 3, 45}
	buf := [disk.PAGE_SIZE]byte{}

	assert.Equal(t, m.ReadPage(0, unsafe.Pointer(&buf)), disk.ErrIOReadExceedFileSize, "Failed to read a page")
	assert.Equal(t, [disk.PAGE_SIZE]byte{}, buf)

	assert.Nil(t, m.WritePage(1, unsafe.Pointer(&data)), "Failed to write a page")
	assert.Nil(t, m.ReadPage(1, unsafe.Pointer(&buf)), "Failed to read a page")
	assert.Equal(t, buf, data)

	assert.Nil(t, m.WritePage(10, unsafe.Pointer(&data)), "Failed to write a page")
	assert.Nil(t, m.ReadPage(10, unsafe.Pointer(&buf)), "Failed to read a page")
	assert.Equal(t, buf, data)

	assert.Nil(t, m.ShutDown())
	assert.Nil(t, os.Remove("test.db"))
}
