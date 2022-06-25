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

package hybrid

import (
	"sync"
	"sync/atomic"
)

// FreeList thread-safe FILO stack
type FreeList struct {
	head atomic.Value
	// for head is nil
	mu      sync.Mutex
	counter uint64
}

/*
Pop first element from front

Notes: The ABA Problem http://15418.courses.cs.cmu.edu/spring2013/article/46.

Thread 0 begins a pop and sees "A" as the top, followed by "B".

Thread 1 begins and completes a pop, returning "A".

Thread 1 begins and completes a push of "D".

Thread 1 pushes "A" back onto the stack and completes.

Thread 0 sees that "A" is on top and returns "A", setting the new top to "B".

Node D is lost.

TODO: in our scenarios, can we ignore the ABA problem?
*/
func (l *FreeList) Pop() (bf *BufferFrame) {
	curHead := l.head.Load()

	for curHead != nil {
		next := curHead.(*BufferFrame).nextFreeBF
		if l.head.CompareAndSwap(curHead, next) {
			bf = curHead.(*BufferFrame)
			bf.nextFreeBF = nil
			atomic.AddUint64(&l.counter, ^uint64(0)) // subtract 1
			// TODO: assert bf is not latched
			// TODO: assert bf state if free
			return bf
		} else {
			if curHead == nil {
				break // return nil
			} else {
				curHead = l.head.Load()
			}
		}
	}

	return nil
}

// Push an element into stack
func (l *FreeList) Push(bf *BufferFrame) {
	// TODO: assert bf state if free
	// TODO: assert bf is not latched
	if l.head.Load() != nil {
		bf.nextFreeBF = l.head.Load().(*BufferFrame)
		for !l.head.CompareAndSwap(bf.nextFreeBF, bf) {
			//log.Println("failed")
			bf.nextFreeBF = l.head.Load().(*BufferFrame)
		}
	} else {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.head.Swap(bf)
	}
	atomic.AddUint64(&l.counter, 1)
}

func (l *FreeList) BatchPush(batchHead, batchTail *BufferFrame, len uint64) {
	batchTail.nextFreeBF = l.head.Load().(*BufferFrame)
	for !l.head.CompareAndSwap(batchTail.nextFreeBF, batchHead) {
		batchTail.nextFreeBF = l.head.Load().(*BufferFrame)
	}
	atomic.AddUint64(&l.counter, len)
}
