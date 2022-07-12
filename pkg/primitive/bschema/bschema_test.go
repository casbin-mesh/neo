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

package bschema

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBSchema_EncodeKey(t *testing.T) {
	rw := NewReaderWriter([]byte("Hello"), []byte("p"))
	dst := rw.EncodeKey()
	assert.Equal(t, 8, len(dst))
}

func TestBSchema_DecodeKey(t *testing.T) {
	rw := NewReaderWriter([]byte("Hello"), []byte("p"))
	dst := rw.EncodeKey()
	assert.Equal(t, 8, len(dst))

	dec := readerWriter{}
	dec.DecodeKey(dst)
	assert.Equal(t, []byte("Hello"), dec.namespace)
	assert.Equal(t, []byte("p"), dec.name)
}
