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

package ast

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDfsIterator_Next(t *testing.T) {
	tree := &BinaryOperationExpr{
		Op: AND_AND,
		L: &BinaryOperationExpr{
			Op: AND_AND,
			L: &Primitive{
				Typ:   INT,
				Value: 1,
			},
			R: &Primitive{
				Typ:   INT,
				Value: 2,
			},
		},
		R: &BinaryOperationExpr{
			Op: AND_AND,
			L: &Primitive{
				Typ:   INT,
				Value: 3,
			},
			R: &Primitive{
				Typ:   INT,
				Value: 4,
			},
		},
	}

	iter := NewDfsIterator(tree)

	// root
	n := iter.Next()
	expr, ok := n.(*BinaryOperationExpr)
	assert.True(t, ok)
	assert.Equal(t, expr.Op, AND_AND)

	// left and
	n = iter.Next()
	expr, ok = n.(*BinaryOperationExpr)
	assert.True(t, ok)
	assert.Equal(t, expr.Op, AND_AND)

	// left 1
	n = iter.Next()
	pri, ok := n.(*Primitive)
	assert.True(t, ok)
	assert.Equal(t, pri.Value, 1)

	// right 2
	n = iter.Next()
	pri, ok = n.(*Primitive)
	assert.True(t, ok)
	assert.Equal(t, pri.Value, 2)

	// right and
	n = iter.Next()
	expr, ok = n.(*BinaryOperationExpr)
	assert.True(t, ok)
	assert.Equal(t, expr.Op, AND_AND)

	// left 3
	n = iter.Next()
	pri, ok = n.(*Primitive)
	assert.True(t, ok)
	assert.Equal(t, pri.Value, 3)

	// right 4
	n = iter.Next()
	pri, ok = n.(*Primitive)
	assert.True(t, ok)
	assert.Equal(t, pri.Value, 4)

	assert.False(t, iter.HasNext())
}

func TestBfsIterator_Next(t *testing.T) {
	tree := &BinaryOperationExpr{
		Op: AND_AND,
		L: &BinaryOperationExpr{
			Op: AND_AND,
			L: &Primitive{
				Typ:   INT,
				Value: 1,
			},
			R: &Primitive{
				Typ:   INT,
				Value: 2,
			},
		},
		R: &BinaryOperationExpr{
			Op: AND_AND,
			L: &Primitive{
				Typ:   INT,
				Value: 3,
			},
			R: &Primitive{
				Typ:   INT,
				Value: 4,
			},
		},
	}

	iter := NewBfsIterator(tree)

	// root
	n := iter.Next()
	expr, ok := n.(*BinaryOperationExpr)
	assert.True(t, ok)
	assert.Equal(t, expr.Op, AND_AND)

	// left and
	n = iter.Next()
	expr, ok = n.(*BinaryOperationExpr)
	assert.True(t, ok)
	assert.Equal(t, expr.Op, AND_AND)

	// right and
	n = iter.Next()
	expr, ok = n.(*BinaryOperationExpr)
	assert.True(t, ok)
	assert.Equal(t, expr.Op, AND_AND)

	// left 1
	n = iter.Next()
	pri, ok := n.(*Primitive)
	assert.True(t, ok)
	assert.Equal(t, pri.Value, 1)

	// left 2
	n = iter.Next()
	pri, ok = n.(*Primitive)
	assert.True(t, ok)
	assert.Equal(t, pri.Value, 2)

	// left 3
	n = iter.Next()
	pri, ok = n.(*Primitive)
	assert.True(t, ok)
	assert.Equal(t, pri.Value, 3)

	// left 4
	n = iter.Next()
	pri, ok = n.(*Primitive)
	assert.True(t, ok)
	assert.Equal(t, pri.Value, 4)

	assert.False(t, iter.HasNext())
}
