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

import "errors"

type path struct {
	node    *Node
	nextIdx int
}

type iterator struct {
	nextNode *Node
	path     []*path
	depth    int
	options  int
}

type NewIteratorConfig struct {
	prefix  Key
	options int
}

func (art *Tree) NewIterator(cfg NewIteratorConfig) *iterator {
	target := art.seekPrefix(cfg.prefix)
	options := traverseOptions(cfg.options)
	path := []*path{{node: target, nextIdx: 0}}
	return &iterator{
		nextNode: target,
		path:     path,
		depth:    0,
		options:  options,
	}
}

func (art *Tree) Iterator(opts ...int) *iterator {
	options := traverseOptions(opts...)
	path := []*path{{node: art.root, nextIdx: 0}}
	return &iterator{
		nextNode: art.root,
		path:     path,
		depth:    0,
		options:  options,
	}
}

func nextChild(childIdx int, children []*Node) (nextChildIdx int, nextNode *Node) {
	for i := childIdx; i < len(children); i++ {
		child := children[i]
		if child != nil {
			return i + 1, child
		}
	}
	return 0, nil
}

func (i *iterator) next() {
	for {
		var nextNode *Node
		curNode := i.path[i.depth].node
		nextChildIdx := i.path[i.depth].nextIdx

		switch curNode.kind {
		case Node4:
			nextChildIdx, nextNode = nextChild(nextChildIdx, curNode.node4().children[:])
		case Node16:
			nextChildIdx, nextNode = nextChild(nextChildIdx, curNode.node16().children[:])
		case Node48:
			nextChildIdx, nextNode = nextChild(nextChildIdx, curNode.node48().children[:])
		case Node256:
			nextChildIdx, nextNode = nextChild(nextChildIdx, curNode.node256().children[:])
		}

		if nextNode == nil {
			if i.depth > 0 { // traverse up
				i.path[i.depth] = nil
				i.depth--
			} else {
				i.nextNode = nil // traverse finish
				return
			}
		} else {
			i.path[i.depth].nextIdx = nextChildIdx
			i.nextNode = nextNode

			// traverse down
			if i.depth+1 >= cap(i.path) {
				newDepthLevel := make([]*path, (i.depth+1)*2)
				copy(newDepthLevel, i.path)
				i.path = newDepthLevel
			}
			i.depth++
			i.path[i.depth] = &path{nextNode, 0}
			return
		}
	}
}

func (i *iterator) HasNext() bool {
	return i.nextNode != nil
}

var (
	ErrNoMoreNodes = errors.New("There are no more nodes in the tree")
)

func (i *iterator) Next() (*Node, error) {
	for i.HasNext() {
		if i.options&TraverseLeaf == TraverseLeaf && i.nextNode.kind == Leaf {
			cur := i.nextNode
			i.next()
			return cur, nil
		} else if i.options&TraverseNode == TraverseNode && i.nextNode.kind != Leaf {
			cur := i.nextNode
			i.next()
			return cur, nil
		}
		i.next()
	}
	return nil, ErrNoMoreNodes
}
