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

package lru

import (
	"github.com/casbin-mesh/neo/pkg/buffer/replacer"
	"github.com/stretchr/testify/assert"
	"math/rand"
)

func verifyVictim(r replacer.Replacer, expected uint64, t assert.TestingT) {
	var res uint64
	assert.True(t, r.Victim(&res))
	assert.Equal(t, expected, res)
}

func generateVictimTest(r replacer.Replacer, size uint64, t assert.TestingT) {
	var res uint64
	assert.False(t, r.Victim(&res), "Should be unable to victim any page")

	// ordering test
	unpinSets := []uint64{rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64()}
	for _, num := range unpinSets {
		assert.Nil(t, r.Unpin(num))
	}
	for _, num := range unpinSets {
		verifyVictim(r, num, t)
	}

	// unpin one
	v := rand.Uint64()
	assert.Nil(t, r.Unpin(v))
	assert.True(t, r.Victim(&res))
	assert.Equal(t, v, res)

	// unpin same values twice
	assert.Nil(t, r.Unpin(uint64(0)))
	assert.Nil(t, r.Unpin(uint64(0)))
	verifyVictim(r, uint64(0), t)

	// unpin test
	for i := uint64(0); i < size; i++ {
		assert.Nil(t, r.Unpin(i))
	}
	testSize := rand.Uint64() % size

	for i := uint64(0); i < testSize; i++ {
		verifyVictim(r, i, t)
	}
	assert.Equal(t, size-testSize, r.Size())
}

func generatePinTest(r replacer.Replacer, size uint64, t assert.TestingT) {
	var res uint64

	v := rand.Uint64()
	// try removing values not exist in the replacer
	assert.Nil(t, r.Pin(v))

	// pin twice
	assert.Nil(t, r.Unpin(v))
	assert.Nil(t, r.Pin(v))
	assert.Nil(t, r.Pin(v))

	assert.Nil(t, r.Unpin(v))
	assert.Nil(t, r.Pin(v))
	assert.False(t, r.Victim(&res), "A pin value victimized")

	// ordering test
	unpinSets := []uint64{rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64()}
	pinnedIndex := rand.Int() % len(unpinSets)
	for _, num := range unpinSets {
		assert.Nil(t, r.Unpin(num))
	}
	assert.Nil(t, r.Pin(unpinSets[pinnedIndex]))
	for i, num := range unpinSets {
		if i != pinnedIndex {
			verifyVictim(r, num, t)
		}
	}

	assert.Equal(t, uint64(0), r.Size())

	// unpin all
	for i := uint64(0); i < size; i++ {
		assert.Nil(t, r.Unpin(i))
	}

	testSize := rand.Uint64() % size
	// pin partial values
	for i := uint64(0); i < testSize/2; i++ {
		assert.Nil(t, r.Pin(i))
	}
	// remove partial values
	for i := testSize / 2; i < testSize; i++ {
		verifyVictim(r, i, t)
	}

}

func generateSizeTest(r replacer.Replacer, size uint64, t assert.TestingT) {
	testSize := rand.Uint64() % size

	// size test
	for i := uint64(0); i < testSize; i++ {
		assert.Equal(t, i, r.Size())
		assert.Nil(t, r.Unpin(i))
	}
	assert.Nil(t, r.Unpin(testSize-1))
	assert.Equal(t, testSize, r.Size())

	// victim values
	for i := uint64(0); i < testSize; i++ {
		verifyVictim(r, i, t)
	}

	assert.Equal(t, uint64(0), r.Size())

	// double size
	for i := uint64(0); i < size; i++ {
		assert.Nil(t, r.Unpin(i))
	}

	assert.Equal(t, size, r.Size())

	for i := size; i < size*2; i++ {
		assert.Equal(t, ErrExceedMaxCap, r.Unpin(i))
	}
}
