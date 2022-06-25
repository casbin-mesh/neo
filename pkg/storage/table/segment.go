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

package table

import "github.com/casbin-mesh/neo/pkg/storage/table/run"

type Segment interface {
	// Smallest returns the smallest key in segment
	Smallest() []byte
	// Biggest returns the biggest key in segment
	Biggest() []byte
	// Search uses binary search to find and return the smallest index i in [0, n) at which f(i) is true,
	// assuming that on the range [0, n), f(i) == true implies f(i+1) == true.
	Search(n int, f func(int) bool) int
}

type cursorOffset struct {
	block uint32
	keyId uint32 // slotId
}

type segmentIndex struct {
	anchorKey     []byte
	cursorOffsets []cursorOffset // the first element which bigger than anchorKey in each the runs
	runSelectors  []byte         // determiners
}

type segment struct {
	runs  []run.Run
	index segmentIndex
	// bloom filter
}
