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
	"github.com/casbin-mesh/neo/pkg/buffer/replacer/lru"
	"github.com/casbin-mesh/neo/pkg/storage/disk/simple"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTinyTest(t *testing.T) {
	// init
	dm := simple.Default()
	assert.Nil(t, dm.Open("test.db"), "Failed to open test.db file")
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
}

func TestSampleTest(t *testing.T) {
	// init
	dm := simple.Default()
	assert.Nil(t, dm.Open("test.db"), "Failed to open test.db file")
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
}
