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

package page

import (
	"github.com/casbin-mesh/neo/pkg/storage/disk"
	"sync"
)

// InvalidPageId first is used to identify whether the current page is valid
const InvalidPageId = uint64(1 << 63)

type Page interface {
	IsDirty() bool
	Latch() Locker
	PinCount() uint64
	DecrPinCount() uint64
	SetPinCount(count uint64)
	SetPageId(pid uint64)
	PageId() uint64
	Data() *[disk.PAGE_SIZE]byte
	SetIsDirty(bool)
	ResetData()
}

type Locker interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type page struct {
	data     [disk.PAGE_SIZE]byte
	pageId   uint64
	pinCount uint64
	isDirty  bool
	latch    sync.RWMutex
}

func (p *page) DecrPinCount() uint64 {
	p.pinCount--
	return p.pinCount
}

func (p *page) SetPageId(pid uint64) {
	p.pageId = pid
}

func (p *page) SetIsDirty(b bool) {
	p.isDirty = b
}

func memset(data *[disk.PAGE_SIZE]byte, value byte) {
	data[0] = value
	for i := 1; i < disk.PAGE_SIZE; i *= 2 {
		copy(data[i:], data[:i])
	}
}

func (p *page) ResetData() {
	//TODO: may rewrite to an assembly implementation
	// check this: https://github.com/tmthrgd/go-memset
	memset(&p.data, 0)
}

func (p *page) SetPinCount(count uint64) {
	p.pinCount = count
}

func (p *page) IsDirty() bool {
	return p.isDirty
}

func (p *page) Latch() Locker {
	return &p.latch
}

func (p *page) PinCount() uint64 {
	return p.pinCount
}

func (p *page) PageId() uint64 {
	return p.pageId
}

func (p *page) Data() *[disk.PAGE_SIZE]byte {
	return &p.data
}

func NewPage(pid uint64) Page {
	return &page{pageId: pid}
}
