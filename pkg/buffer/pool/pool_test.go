// Copyright 2022 The casbin-neo Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pool

import (
	"fmt"
	"github.com/casbin-mesh/neo/pkg/buffer/replacer/lru"
	"github.com/casbin-mesh/neo/pkg/storage/disk/simple"
	"github.com/casbin-mesh/neo/pkg/storage/page"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const filename = "test.db"

func TestTinyTest(t *testing.T) {
	// init
	dm := simple.Default()
	assert.Nil(t, dm.Open(filename), "Failed to open test.db file")
	size := uint64(3)
	r := lru.NewLRUReplacer(size)
	bufferPool := NewBufferPool(Options{DiskManager: dm, Replacer: r, Cap: size, Len: size})
	var pageId uint64
	page, err := bufferPool.NewPage(nil, &pageId)
	assert.Nil(t, err, "Failed to allocate a new page")
	assert.NotNil(t, page, "Page is nil")
	assert.Equal(t, pageId, uint64(0), "PageId should start from 0")

	pageData := "Hello World!"
	// write data into page
	copy(page.Data()[:], pageData)

	for i := uint64(1); i < size; i++ {
		page, err = bufferPool.NewPage(nil, &pageId)
		assert.Nil(t, err, "Failed to allocate a new page at %d", i)
		assert.NotNil(t, page, "Page is nil at %d", i)
	}
	// unpin all pages
	for i := uint64(0); i < size; i++ {
		ok, err := bufferPool.UnpinPage(nil, i, true)
		assert.Nil(t, err, "Failed to unpin page at %d", i)
		assert.True(t, ok, "It should be ok")
	}

	// allocate new pages
	for i := uint64(0); i < size; i++ {
		page, err = bufferPool.NewPage(nil, &pageId)
		assert.Nil(t, err, "Failed to allocate a new page at %d", i)
		assert.NotNil(t, page, "Page is nil at %d", i)
	}
	os.Remove(filename)
}

func TestSampleTest(t *testing.T) {
	// init
	dm := simple.Default()
	assert.Nil(t, dm.Open(filename), "Failed to open test.db file")
	size := uint64(10)
	r := lru.NewLRUReplacer(size)
	bufferPool := NewBufferPool(
		Options{
			DiskManager: dm,
			Replacer:    r,
			Cap:         size,
			Len:         size,
		})
	// tests
	var pageId uint64
	page, err := bufferPool.NewPage(nil, &pageId)
	assert.Nil(t, err, "Failed to allocate a new page")
	assert.NotNil(t, page, "Page is nil")
	assert.Equal(t, pageId, uint64(0), "PageId should start from 0")

	pageData := "Hello World!"
	// write data into page
	copy(page.Data()[:], pageData)

	// fill up the buffer pool.
	for i := uint64(1); i < size; i++ {
		page, err = bufferPool.NewPage(nil, &pageId)
		assert.Nil(t, err, "Failed to allocate a new page at %d", i)
		assert.NotNil(t, page, "Page is nil at %d", i)
	}

	// should not be able to create any new pages.
	for i := size; i < size*2; i++ {
		page, _ = bufferPool.NewPage(nil, &pageId)
		assert.Nil(t, page, "It should not be able to create page at %d", i)
	}

	// unpin all pages
	for i := uint64(0); i < size; i++ {
		ok, err := bufferPool.UnpinPage(nil, i, true)
		assert.Nil(t, err, "Failed to unpin page at %d", i)
		assert.True(t, ok, "It should be ok")
	}
	// create 9 pages, and remain one slot for further tests
	for i := uint64(0); i < size-1; i++ {
		page, err = bufferPool.NewPage(nil, &pageId)
		assert.Nil(t, err, "Failed to allocate a new page at %d", i)
		assert.NotNil(t, page, "Page is nil at %d", i)
	}
	// should be able to retrieve the page from disk
	page0, err := bufferPool.FetchPage(nil, uint64(0))
	assert.Nil(t, err, "Failed to retrieve page 0")
	assert.Equal(t, string(page0.Data()[:len(pageData)]), pageData)

	ok, err := bufferPool.UnpinPage(nil, uint64(0), false)
	assert.True(t, ok, "Failed to unpin page 0")
	assert.Nil(t, err, "Failed to unpin page 0")
	os.Remove(filename)
}

func defaultPool(t *testing.T, size uint64) BufferPoolManager {
	dm := simple.Default()
	assert.Nil(t, dm.Open(filename), "Failed to open test.db file")
	r := lru.NewLRUReplacer(size)
	bufferPool := NewBufferPool(
		Options{
			DiskManager: dm,
			Replacer:    r,
			Cap:         size,
			Len:         size,
		})
	return bufferPool
}

func TestPool_NewPage(t *testing.T) {
	size := uint64(10)
	p := defaultPool(t, size)

	// track lru replacer data
	var insertedPgIdsFIFO []uint64
	// fill pool
	var pageId uint64
	for i := uint64(0); i < size; i++ {
		pg, err := p.NewPage(nil, &pageId)
		assert.NotNil(t, pg)
		assert.Nil(t, err)
		insertedPgIdsFIFO = append(insertedPgIdsFIFO, pageId)
	}

	// It should be unable to create any new pages
	for i := uint64(0); i < size; i++ {
		pg, err := p.NewPage(nil, &pageId)
		assert.Equal(t, ErrNoAvailablePage, err)
		assert.Nil(t, pg)
	}

	UnpinSomePages := func() {
		// unpin some pages
		for i := uint64(0); i < size/2; i++ {
			ok, err := p.UnpinPage(nil, insertedPgIdsFIFO[i], false)
			assert.Nil(t, err)
			assert.True(t, ok)
		}
	}

	CreateNewPages := func() {
		// unpinned pages will be evicted
		for i := uint64(0); i < size/2; i++ {
			pg, err := p.NewPage(nil, &pageId)
			assert.NotNil(t, pg)
			assert.Nil(t, err)
			insertedPgIdsFIFO[i] = pageId
		}
	}

	ShouldUnableCreate := func() {
		// It should be unable to create any new pages
		for i := uint64(0); i < size; i++ {
			pg, err := p.NewPage(nil, &pageId)
			assert.Equal(t, ErrNoAvailablePage, err)
			assert.Nil(t, pg)
		}
	}

	tasks := []func(){UnpinSomePages, CreateNewPages, ShouldUnableCreate, UnpinSomePages, CreateNewPages, ShouldUnableCreate}

	for _, task := range tasks {
		task()
	}

	os.Remove(filename)
}

func NewPageHelper(p BufferPoolManager, data string, t assert.TestingT) {
	var pgId uint64
	pg, err := p.NewPage(nil, &pgId)
	assert.NotNil(t, pg)
	assert.Nil(t, err)
	copy(pg.Data()[:], data)
}

func TestPool_UnpinPage(t *testing.T) {
	size := uint64(10)
	p := defaultPool(t, size)

	// create new pages
	for i := uint64(0); i < size; i++ {
		NewPageHelper(p, fmt.Sprintf("inserted data %d", i), t)
	}

	// unpin created pages
	for i := uint64(0); i < size; i++ {
		ok, err := p.UnpinPage(nil, i, true)
		assert.Nil(t, err)
		assert.True(t, ok)
	}

	// create more new pages
	for i := size; i < size*2; i++ {
		NewPageHelper(p, fmt.Sprintf("inserted data %d", i), t)
	}

	// unpin created pages
	for i := size; i < size*2; i++ {
		ok, err := p.UnpinPage(nil, i, true)
		assert.Nil(t, err)
		assert.True(t, ok)
	}

	// check inserted data and update data
	for i := uint64(0); i < size; i++ {
		expected := fmt.Sprintf("inserted data %d", i)
		pg, err := p.FetchPage(nil, i)
		assert.Nil(t, err)
		assert.NotNil(t, pg)
		assert.Equal(t, []byte(expected), pg.Data()[:len(expected)])
		copy(pg.Data()[:], fmt.Sprintf("updated data %d", i))
	}

	// unpin pages
	for i := uint64(0); i < size; i++ {
		ok, err := p.UnpinPage(nil, i, i%2 == 0)
		assert.Nil(t, err)
		assert.True(t, ok)
	}

	// should be able to create new pages
	for i := uint64(0); i < size; i++ {
		var pgId uint64
		pg, err := p.NewPage(nil, &pgId)
		assert.NotNil(t, pg)
		assert.Nil(t, err)
		ok, err := p.UnpinPage(nil, pgId, true)
		assert.Nil(t, err)
		assert.True(t, ok)
	}

	// test updated data
	for i := uint64(0); i < size; i++ {
		if i%2 == 0 {
			// updated
			expected := fmt.Sprintf("updated data %d", i)
			pg, err := p.FetchPage(nil, i)
			assert.Nil(t, err)
			assert.NotNil(t, pg)
			assert.Equal(t, []byte(expected), pg.Data()[:len(expected)])
		} else {
			// discarded
			expected := fmt.Sprintf("inserted data %d", i)
			pg, err := p.FetchPage(nil, i)
			assert.Nil(t, err)
			assert.NotNil(t, pg)
			assert.Equal(t, []byte(expected), pg.Data()[:len(expected)])
		}

	}

	os.Remove(filename)
}

func TestPool_FetchPage(t *testing.T) {
	size := uint64(10)
	p := defaultPool(t, size)
	// track lru replacer data
	var insertedPgIdsFIFO []uint64
	insertedData := "Hello World!"
	// fill pool
	var pageId uint64
	var pages []page.Page
	for i := uint64(0); i < size; i++ {
		pg, err := p.NewPage(nil, &pageId)
		assert.NotNil(t, pg)
		assert.Nil(t, err)
		pages = append(pages, pg)
		insertedPgIdsFIFO = append(insertedPgIdsFIFO, pageId)
		copy(pg.Data()[:], insertedData)
	}

	// should hit the pool cache
	for i := uint64(0); i < size; i++ {
		pg, err := p.FetchPage(nil, insertedPgIdsFIFO[i])
		assert.NotNil(t, pg)
		assert.Nil(t, err)
		assert.Equal(t, pages[i], pg)
		assert.Equal(t, pg.Data()[:len(insertedData)], []byte(insertedData))
		// unpin twice
		ok, err := p.UnpinPage(nil, insertedPgIdsFIFO[i], true)
		assert.True(t, ok)
		assert.Nil(t, err)
		ok, err = p.UnpinPage(nil, insertedPgIdsFIFO[i], true)
		assert.True(t, ok)
		assert.Nil(t, err)
	}

	// create new page and unpin
	for i := uint64(0); i < size; i++ {
		pg, err := p.NewPage(nil, &pageId)
		assert.NotNil(t, pg)
		assert.Nil(t, err)
		ok, err := p.UnpinPage(nil, pageId, true)
		assert.True(t, ok)
		assert.Nil(t, err)
	}

	// should fetch from disk
	for i := uint64(0); i < size; i++ {
		pg, err := p.FetchPage(nil, insertedPgIdsFIFO[i])
		assert.NotNil(t, pg)
		assert.Nil(t, err)
		assert.Equal(t, pages[i], pg)
		assert.Equal(t, pg.Data()[:len(insertedData)], []byte(insertedData))
		// fill data and discard
		copy(pg.Data()[:], "updated")
		ok, err := p.UnpinPage(nil, insertedPgIdsFIFO[i], false)
		assert.True(t, ok)
		assert.Nil(t, err)
	}

	// create new page and unpin
	for i := uint64(0); i < size; i++ {
		pg, err := p.NewPage(nil, &pageId)
		assert.NotNil(t, pg)
		assert.Nil(t, err)
		ok, err := p.UnpinPage(nil, pageId, true)
		assert.True(t, ok)
		assert.Nil(t, err)
	}

	// should fetch from disk
	for i := uint64(0); i < size; i++ {
		pg, err := p.FetchPage(nil, insertedPgIdsFIFO[i])
		assert.NotNil(t, pg)
		assert.Nil(t, err)
		assert.Equal(t, pages[i], pg)
		assert.Equal(t, pg.Data()[:len(insertedData)], []byte(insertedData))
		ok, err := p.UnpinPage(nil, insertedPgIdsFIFO[i], false)
		assert.True(t, ok)
		assert.Nil(t, err)
	}

	os.Remove(filename)
}

func TestPool_DeletePage(t *testing.T) {
	size := uint64(10)
	p := defaultPool(t, size)

	// track lru replacer data
	var insertedPgIdsFIFO []uint64
	insertedData := "Hello World!"
	// fill pool
	var pageId uint64
	var pages []page.Page
	for i := uint64(0); i < size; i++ {
		pg, err := p.NewPage(nil, &pageId)
		assert.NotNil(t, pg)
		assert.Nil(t, err)
		pages = append(pages, pg)
		insertedPgIdsFIFO = append(insertedPgIdsFIFO, pageId)
		copy(pg.Data()[:], insertedData)
	}

	// should hit the pool cache
	for i := uint64(0); i < size; i++ {
		pg, err := p.FetchPage(nil, insertedPgIdsFIFO[i])
		assert.NotNil(t, pg)
		assert.Nil(t, err)
		assert.Equal(t, pages[i], pg)
		assert.Equal(t, pg.Data()[:len(insertedData)], []byte(insertedData))
		// unpin twice
		ok, err := p.UnpinPage(nil, insertedPgIdsFIFO[i], true)
		assert.True(t, ok)
		assert.Nil(t, err)
		ok, err = p.UnpinPage(nil, insertedPgIdsFIFO[i], true)
		assert.True(t, ok)
		assert.Nil(t, err)
	}

	// create new page and unpin
	for i := uint64(0); i < size; i++ {
		pg, err := p.NewPage(nil, &pageId)
		assert.NotNil(t, pg)
		assert.Nil(t, err)
		ok, err := p.UnpinPage(nil, pageId, true)
		assert.True(t, ok)
		assert.Nil(t, err)
	}

	// should fetch from disk
	for i := uint64(0); i < size; i++ {
		pg, err := p.FetchPage(nil, insertedPgIdsFIFO[i])
		assert.NotNil(t, pg)
		assert.Nil(t, err)
		assert.Equal(t, pages[i], pg)
		assert.Equal(t, pg.Data()[:len(insertedData)], []byte(insertedData))
	}

	// should be unable to delete pages, due to pin count is not 0
	for i := uint64(0); i < size; i++ {
		ok, err := p.DeletePage(nil, insertedPgIdsFIFO[i])
		assert.False(t, ok)
		assert.Equal(t, ErrPinCountZero, err)
	}

	// unpin and delete
	for i := uint64(0); i < size; i++ {
		ok, err := p.UnpinPage(nil, insertedPgIdsFIFO[i], false)
		assert.True(t, ok)
		assert.Nil(t, err)
		ok, err = p.DeletePage(nil, insertedPgIdsFIFO[i])
		assert.True(t, ok)
		assert.Nil(t, err)
	}

	os.Remove(filename)
}

func TestPool_UnpinPage_Dirty(t *testing.T) {
	size := uint64(1)
	p := defaultPool(t, size)
	insertedData := "Hello World!"

	// create page 0
	var pageId0 uint64
	pg0, err := p.NewPage(nil, &pageId0)
	assert.Nil(t, err)
	assert.NotNil(t, pg0)
	copy(pg0.Data()[:], insertedData)
	ok, err := p.UnpinPage(nil, pageId0, true)
	assert.True(t, ok)
	assert.Nil(t, err)

	// should be dirty
	pg0, err = p.FetchPage(nil, pageId0)
	assert.Nil(t, err)
	assert.NotNil(t, pg0)
	assert.True(t, pg0.IsDirty())
	assert.Equal(t, []byte(insertedData), pg0.Data()[:len(insertedData)])
	ok, err = p.UnpinPage(nil, pageId0, false)
	assert.True(t, ok)
	assert.Nil(t, err)

	// should be dirty
	pg0, err = p.FetchPage(nil, pageId0)
	assert.Nil(t, err)
	assert.NotNil(t, pg0)
	assert.True(t, pg0.IsDirty())
	ok, err = p.UnpinPage(nil, pageId0, false)
	assert.True(t, ok)
	assert.Nil(t, err)

	insertedData2 := "Hello Page1!"
	// new page1
	var pageId1 uint64
	pg1, err := p.NewPage(nil, &pageId1)
	assert.Nil(t, err)
	assert.NotNil(t, pg1)
	assert.False(t, pg1.IsDirty())

	// write data into page1
	copy(pg0.Data()[:], insertedData2)
	ok, err = p.UnpinPage(nil, pageId1, true)
	assert.True(t, ok)
	assert.Nil(t, err)
	// should be able to delete page1
	ok, err = p.DeletePage(nil, pageId1)
	assert.True(t, ok)
	assert.Nil(t, err)

	// fetch page0 and verify data
	pg0, err = p.FetchPage(nil, pageId0)
	assert.Nil(t, err)
	assert.NotNil(t, pg0)
	assert.False(t, pg0.IsDirty())
	assert.Equal(t, []byte(insertedData), pg0.Data()[:len(insertedData)])
	ok, err = p.UnpinPage(nil, pageId0, false)
	assert.True(t, ok)
	assert.Nil(t, err)

	os.Remove(filename)
}

// TODO: add concurrency tests
