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
	"container/list"
	"errors"
	"github.com/casbin-mesh/neo/pkg/buffer/replacer"
	"github.com/casbin-mesh/neo/pkg/storage/disk"
	"github.com/casbin-mesh/neo/pkg/storage/page"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Ctx TODO
type Ctx interface {
}

type BufferPoolManager interface {
	FetchPage(ctx Ctx, pageId uint64, opts ...*FetchPageOption) (page.Page, error)
	UnpinPage(ctx Ctx, pageId uint64, dirty bool) (bool, error)
	NewPage(ctx Ctx, pageId *uint64, opts ...*NewPageOption) (page.Page, error)
	DeletePage(ctx Ctx, pageId uint64) (bool, error)
	FlushAll(ctx Ctx) error
}

type Options struct {
	DiskManager disk.BasicManager
	Replacer    replacer.Replacer
	Cap         uint64
	Len         uint64
}

type Frame uint64
type PageID uint64

type pool struct {
	dm disk.BasicManager
	r  replacer.Replacer

	nextPageId uint64

	//  max number of tracked pages
	pagesMu    sync.RWMutex
	pages      []page.Page
	cap        uint64
	freelistMu sync.RWMutex
	freelist   []Frame

	// TODO: should we use multi-level pageTable instead?
	// pageTable for tracks cached in-memory pages.
	// type of element (Frame), index of pages
	pageTable   map[PageID]Frame
	pageTableMu sync.RWMutex

	// TODO: should we track the deleted pages (the yet not deallocated page)
	// deletedPageIds store deletedPageIds, for reuse deleted pageIds
	// type of element (PageID)
	deletedPageIds   *list.List
	deletedPageIdsMu sync.RWMutex
}

// FetchPage return a requested page.
//
// If the requested page in memory, pin it, return page.
//
// Else we pick an available frame from freelist or replacer,
// remove the origin page mapping of pageTable,
// and insert a new page2frame mapping into pageTable,
func (p *pool) FetchPage(ctx Ctx, pageId uint64, opts ...*FetchPageOption) (pg page.Page, err error) {
	// TODO: handling option later

	frameId, ok := p.pageTable[PageID(pageId)]
	// If hit the cache
	if ok {
		if err = p.r.Pin(uint64(frameId)); err != nil {
			return nil, err
		}
		pg = p.pages[frameId]
		pg.IncrPinCount()
		return pg, nil
	}

	// pick an available frameId
	evictedFrameId, err := p.findOneAvailableFrame(nil)
	if err != nil {
		return nil, err
	}

	pg = p.pages[evictedFrameId]
	if err = p.r.Pin(uint64(evictedFrameId)); err != nil {
		return nil, err
	}
	p.pageTableMu.Lock()
	defer p.pageTableMu.Unlock()
	p.pageTable[PageID(pageId)] = evictedFrameId
	// clean up
	pg.SetPageId(pageId)
	pg.ResetData()
	pg.SetIsDirty(false)
	pg.SetPinCount(1)
	// read from disk
	err = p.dm.ReadPage(pageId, unsafe.Pointer(pg.Data()))
	if err != nil {
		return nil, err
	}

	return pg, nil
}

func (p *pool) UnpinPage(ctx Ctx, pageId uint64, dirty bool) (bool, error) {
	frame, ok := p.pageTable[PageID(pageId)]
	if !ok {
		return false, nil
	}
	pg := p.pages[frame]

	if pg.PinCount() == 0 {
		return false, nil
	}
	if dirty {
		pg.SetIsDirty(dirty)
	}
	pg.DecrPinCount()
	if pg.PinCount() == 0 {
		if err := p.r.Unpin(uint64(frame)); err != nil {
			return false, err
		}
	}
	return true, nil
}

var (
	ErrNoAvailablePage = errors.New("failed to find an available page")
	ErrPinCountZero    = errors.New("page pin count is not zero")
)

// findVictimPage returns an available Frame.
//
// if it returns a deleted page (tracked by deletedPageIds)
//
// First it will check the pool, if there is a free page able to be found
// then it will check the replacer, if there is a page can be victim(replaced).
func (p *pool) findOneAvailableFrame(opts *FindOneAvailablePageOption) (Frame, error) {
	// TODO: allow user to decide the ordering of reusing / allocating  / replacing

	// First it will check the pool, if there is a free page able to be found
	// TODO: handle the growing of freelist
	p.freelistMu.Lock()
	if len(p.freelist) != 0 {
		first := p.freelist[0]
		p.freelist = p.freelist[1:]
		p.freelistMu.Unlock()

		return first, nil
	}
	p.freelistMu.Unlock()

	// If above operations failed, then we try to victim a page form replacer
	var victim uint64
	if p.r.Victim(&victim) {
		p.pagesMu.RLock()
		defer p.pagesMu.RUnlock()
		pg := p.pages[victim]
		delete(p.pageTable, PageID(pg.PageId()))

		if pg.IsDirty() {
			// TODO: batch or epoch base flush
			// for now, just flush per operation (fsync)
			if err := p.dm.WritePage(pg.PageId(), unsafe.Pointer(pg.Data())); err != nil {
				return 0, err
			}
		}

		return Frame(victim), nil
	}

	return 0, ErrNoAvailablePage
}

// getNextPageID returns next page Id
func (p *pool) getNextPageID() PageID {
	return PageID(atomic.AddUint64(&p.nextPageId, 1) - 1)
}

// NewPage returns a new page.
// If it cannot find an available page form the replacer either the pool,
// which means all pages were pinned, then it will return an error
func (p *pool) NewPage(ctx Ctx, pageId *uint64, opts ...*NewPageOption) (page.Page, error) {

	//TODO: for now, options is nil
	frame, err := p.findOneAvailableFrame(nil)
	if err != nil {
		return nil, err
	}
	var pg page.Page

	// Reuse the un-deallocated page.Page

	// Allocate a new page
	// TODO: use pre-allocation to amortize the overhead of memory allocations
	pg = p.pages[frame]
	nextPage := uint64(p.getNextPageID())
	*pageId = nextPage
	pg.ResetData()
	pg.SetIsDirty(false)
	pg.SetPinCount(1)
	pg.SetPageId(nextPage)

	// Register page to pageTable
	p.pageTableMu.Lock()
	p.pageTable[PageID(nextPage)] = frame
	p.pageTableMu.Unlock()
	return pg, nil
}

func (p *pool) DeletePage(ctx Ctx, pageId uint64) (bool, error) {
	p.pageTableMu.Lock()
	defer p.pageTableMu.Unlock()
	frame, ok := p.pageTable[PageID(pageId)]
	pg := p.pages[frame]
	if ok {
		if pg.PinCount() != 0 {
			return false, ErrPinCountZero
		}
		pg.ResetData()
		// deregister page
		delete(p.pageTable, PageID(pageId))
		p.deletedPageIdsMu.Lock()
		defer p.deletedPageIdsMu.Unlock()
		p.deletedPageIds.PushFront(pg)
	}
	// if page does not exist, return true
	return true, nil
}

func (p *pool) FlushAll(ctx Ctx) error {
	panic("implement me")
}

func (p *pool) InitPages(len uint64) {
	for i := uint64(0); i < len; i++ {
		p.freelist[i] = Frame(i)
		p.pages = append(p.pages, page.NewPage(i))
	}
}

// NewBufferPool
// TODO: pre-allocation
func NewBufferPool(o Options) BufferPoolManager {
	p := &pool{
		dm:             o.DiskManager,
		r:              o.Replacer,
		cap:            o.Cap,
		freelist:       make([]Frame, o.Len),
		pageTable:      make(map[PageID]Frame),
		deletedPageIds: list.New(),
	}
	p.InitPages(o.Len)

	return p
}
