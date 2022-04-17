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

package simple

import (
	"github.com/casbin-mesh/neo/pkg/storage/disk"
	"log"
	"os"
	"unsafe"
)

type DiskManager struct {
	file *os.File
	disk.Options
}

func (m *DiskManager) Open(filename string, opts ...disk.Option) (err error) {
	// Apply additional options
	opt := m.Options.Clone()
	for _, o := range opts {
		err = o(&opt)
		if err != nil {
			return
		}
	}

	var file *os.File
	// Open the target file
	file, err = os.OpenFile(filename, m.Options.Flag, m.Options.Perm)
	if err != nil {
		return
	}
	m.file = file

	return
}

func (m DiskManager) ShutDown() (err error) {
	if err = m.file.Close(); err != nil {
		return
	}
	return
}

func (m DiskManager) WritePage(pageId uint64, p unsafe.Pointer) (err error) {
	offset := int64(pageId) * disk.PAGE_SIZE
	if _, err = m.file.WriteAt(unsafe.Slice((*byte)(p), disk.PAGE_SIZE), offset); err != nil {
		return
	}
	if err = m.file.Sync(); err != nil {
		return
	}
	return
}

func (m DiskManager) ReadPage(pageId uint64, p unsafe.Pointer) error {
	offset := int64(pageId) * disk.PAGE_SIZE
	fi, err := m.file.Stat()
	if offset+disk.PAGE_SIZE > fi.Size() {
		return disk.ErrIOReadExceedFileSize
	}
	at, err := m.file.ReadAt(unsafe.Slice((*byte)(p), disk.PAGE_SIZE), offset)
	if at < disk.PAGE_SIZE {
		log.Println("I/O warning: read less than page size")
	}
	if err != nil {
		return err
	}
	return nil
}

func New(opts ...disk.Option) (*DiskManager, error) {
	opt := disk.Options{}
	for _, o := range opts {
		err := o(&opt)
		if err != nil {
			return nil, err
		}
	}

	return &DiskManager{
		Options: opt,
	}, nil
}

func Default() *DiskManager {
	return &DiskManager{
		Options: disk.DefaultOptions,
	}
}
