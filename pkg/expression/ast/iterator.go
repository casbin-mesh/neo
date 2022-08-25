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

type Iterator interface {
	Next() Evaluable
	HasNext() bool
}

type stack struct {
	node  Evaluable
	index int
}

type dfsIterator struct {
	nextNode Evaluable
	cursor   int
	stack    []stack
}

func (b *dfsIterator) Next() Evaluable {
	if !b.HasNext() {
		return nil
	}
	cur := b.nextNode
	b.next()
	return cur
}

func (b *dfsIterator) next() {
	var nextNode Evaluable
	for {
		nextIdx := -1
		curNode := b.stack[b.cursor].node
		curIdx := b.stack[b.cursor].index

		nextIdx, nextNode = b.nextChild(curNode, curIdx)

		if nextNode == nil {
			if b.cursor > 0 {
				b.cursor--
			} else {
				b.nextNode = nil
				return
			}
		} else {
			b.stack[b.cursor].index = nextIdx
			b.nextNode = nextNode
			if b.cursor+1 >= cap(b.stack) {
				newStack := make([]stack, b.cursor+10)
				copy(newStack, b.stack)
				b.stack = newStack
			}
			b.cursor++
			b.stack[b.cursor] = stack{nextNode, -1}
			return
		}
	}
}

func (b *dfsIterator) nextChild(cur Evaluable, idx int) (int, Evaluable) {
	if idx == -1 {
		idx = 0
	}
	for i := idx; i < cur.childrenLen(); i++ {
		child := cur.getChildAt(i)
		if child != nil {
			return i + 1, child
		}
	}
	return 0, nil
}

func (b *dfsIterator) HasNext() bool {
	return b.nextNode != nil
}

func NewDfsIterator(root Evaluable) Iterator {
	return &dfsIterator{
		nextNode: root,
		stack:    []stack{{root, -1}},
	}
}

type bfsIterator struct {
	queue []stack
}

func (b *bfsIterator) Next() Evaluable {
	if !b.HasNext() {
		return nil
	}
	var head *stack

	head = &b.queue[0]
	curNode := head.node
	curIndex := head.index

	if curIndex < curNode.childrenLen() {
		for i := curIndex; i < curNode.childrenLen(); i++ {
			nextNode := curNode.getChildAt(i)
			if nextNode != nil {
				b.queue = append(b.queue, stack{nextNode, -1})
			}
		}
		head.index = curNode.childrenLen()
	}

	b.queue = b.queue[1:]
	return curNode
}

func (b *bfsIterator) nextChild(cur Evaluable, idx int) (int, Evaluable) {
	if idx == -1 {
		idx = 0
	}
	for i := idx; i < cur.childrenLen(); i++ {
		child := cur.getChildAt(i)
		if child != nil {
			return i + 1, child
		}
	}
	return 0, nil
}

func (b *bfsIterator) HasNext() bool {
	return len(b.queue) != 0
}

func NewBfsIterator(root Evaluable) Iterator {
	return &bfsIterator{
		queue: []stack{{root, -1}},
	}
}
