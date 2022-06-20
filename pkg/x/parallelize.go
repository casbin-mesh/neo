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

package x

import (
	"runtime"
	"sync"
)

// ParallelRange Parallelize execution. it blocks the current thread.
func ParallelRange(n uint64, worker func(begin, end uint64)) {
	threads := uint64(runtime.NumCPU())
	blockSize := n / threads
	wg := sync.WaitGroup{}
	for i := uint64(0); i < threads; i++ {
		wg.Add(1)
		begin, end := i*blockSize, i*blockSize+blockSize
		if i == threads-1 {
			end = n
		}

		go func(b, e uint64) {
			worker(b, e)
			wg.Done()
		}(begin, end)
	}

	wg.Wait()
}
