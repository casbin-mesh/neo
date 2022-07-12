// Copyright 2022 The casbin-mesh Authors. All Rights Reserved.
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

package btuple

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewHeader(t *testing.T) {
	h := NewHeader(SmallValueType, 8)
	buf := make([]byte, 8)
	h.writeTo(buf)

	decodeTest := header{}
	decodeTest.decode(buf)
	assert.Equal(t, *h, decodeTest)
}
