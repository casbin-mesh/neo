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

package engine

var (
	DefaultBaseOptions   = &BaseOptions{sessionId: nil}
	DefaultInsertOptions = &InsertOptions{DefaultBaseOptions}
	DefaultDeleteOptions = &DeleteOptions{DefaultBaseOptions}
)

type BaseOptions struct {
	sessionId *string
}

type InsertOptions struct {
	*BaseOptions
}

func (io *InsertOptions) Merge(another *InsertOptions) *InsertOptions {
	io.BaseOptions.Merge(another.BaseOptions)
	return io
}

func MergeInsertOptions(opts ...*InsertOptions) *InsertOptions {
	if len(opts) == 0 {
		return DefaultInsertOptions
	}
	if len(opts) == 1 {
		return opts[0]
	}
	merged := &InsertOptions{}
	for _, opt := range opts {
		merged.Merge(opt)
	}
	return merged
}

type UpdateOptions struct {
	*BaseOptions
}

type EnforceOptions struct {
	*BaseOptions
}

type FindOptions struct {
	*BaseOptions
}

type DeleteOptions struct {
	*BaseOptions
}

func (io *DeleteOptions) Merge(another *DeleteOptions) *DeleteOptions {
	io.BaseOptions.Merge(another.BaseOptions)
	return io
}

func MergeDeleteOptions(opts ...*DeleteOptions) *DeleteOptions {
	if len(opts) == 0 {
		return DefaultDeleteOptions
	}
	if len(opts) == 1 {
		return opts[0]
	}
	merged := &DeleteOptions{}
	for _, opt := range opts {
		merged.Merge(opt)
	}
	return merged
}

func (bo *BaseOptions) AutoCommit() bool {
	return bo.sessionId == nil
}

func (bo *BaseOptions) Merge(another *BaseOptions) {
	if another.sessionId != nil {
		bo.sessionId = another.sessionId
	}
}

func MergeBaseOptions(opts ...*BaseOptions) *BaseOptions {
	if len(opts) == 0 {
		return DefaultBaseOptions
	}
	if len(opts) == 1 {
		return opts[0]
	}
	merged := &BaseOptions{}
	for _, opt := range opts {
		merged.Merge(opt)
	}
	return merged
}
