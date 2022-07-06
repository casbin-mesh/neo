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

package art

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type LeftmostSet struct {
	nodeFactor func() inode[Value]
	expected   node[Value]
}

func TestLeftmost(t *testing.T) {
	l := &leaf[Value]{key: Key("a key"), value: Value("value")}
	assert.Equal(t, l, l.leftmost())

	sets := []LeftmostSet{
		{
			nodeFactor: func() inode[Value] {
				return &node4[Value]{}
			},
			expected: l,
		},
		{
			nodeFactor: func() inode[Value] {
				return &node16[Value]{}
			},
			expected: l,
		},
		{
			nodeFactor: func() inode[Value] {
				return &node48[Value]{}
			},
			expected: l,
		},
		{
			nodeFactor: func() inode[Value] {
				return &node256[Value]{}
			},
			expected: l,
		},
	}
	var child node[Value]
	for _, set := range sets {
		// leaf level
		child = l

		// level + 1
		n := set.nodeFactor()
		n.addChild('a', child)
		child = &inner[Value]{node: n}

		// level + 1
		nn := set.nodeFactor()
		nn.addChild('a', child)
		upper := &inner[Value]{node: n}

		leftmost := upper.leftmost()
		assert.NotNil(t, leftmost)
		assert.Equal(t, set.expected, leftmost)
	}

}
