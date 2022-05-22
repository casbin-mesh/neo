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

const (
	// TraverseLeaf Iterate only over Leaf nodes.
	TraverseLeaf = 1

	// TraverseNode Iterate only over non-Leaf nodes.
	TraverseNode = 2

	// TraverseAll Iterate over all nodes in the tree.
	TraverseAll = TraverseLeaf | TraverseNode
)

type Callback func(node *Node) bool

func traverseOptions(opts ...int) int {
	options := 0
	for _, opt := range opts {
		options |= opt
	}
	options &= TraverseAll
	if options == 0 {
		// By default filter only leafs
		options = TraverseLeaf
	}

	return options
}

func traverseFilter(options int, callback Callback) Callback {
	if options == TraverseAll {
		return callback
	}

	return func(node *Node) bool {
		if options&TraverseLeaf == TraverseLeaf && node.kind == Leaf {
			return callback(node)
		} else if options&TraverseNode == TraverseNode && node.kind != Leaf {
			return callback(node)
		}
		return true
	}
}

func (art *Tree) Traversal(callback Callback, opts ...int) {
	options := traverseOptions(opts...)
	art.recursiveTraverse(art.root, traverseFilter(options, callback))
}

func (art *Tree) recursiveTraverse(cur *Node, callback Callback) {
	if cur == nil {
		return
	}
	if !callback(cur) {
		return
	}
	switch cur.kind {
	case Node4:
		art.childrenTraverse(callback, cur.node4().children[:]...)
	case Node16:
		art.childrenTraverse(callback, cur.node16().children[:]...)
	case Node48:
		art.childrenTraverse(callback, cur.node48().children[:]...)
	case Node256:
		art.childrenTraverse(callback, cur.node256().children[:]...)
	}
}

func (art *Tree) childrenTraverse(callback Callback, children ...*Node) {
	for _, child := range children {
		if child != nil {
			art.recursiveTraverse(child, callback)
		}
	}
}

func (art *Tree) seekPrefix(partial Key) *Node {
	cur := art.root
	if partial == nil {
		return art.root
	}
	depth := uint32(0)
	for cur != nil {
		if cur.IsLeaf() {
			if cur.Leaf().PartialMatch(partial, depth) || int(depth) == len(partial) {
				return cur
			}
			// partial match failed
			break
		}

		n := cur.node()
		if n.partialLen > 0 {
			prefixMismatchedIdx := cur.prefixMismatch(partial, depth)
			if depth+prefixMismatchedIdx == uint32(len(partial)) {
				return cur
			}
			if prefixMismatchedIdx == 0 {
				// nis match
				break
			}
			depth += uint32(n.partialLen)
		} else {
			if int(depth) == len(partial) {
				return cur
			}
		}
		next := cur.findChild(partial.At(int(depth)))
		if *next == nil {
			break
		}
		cur = *next
		depth++
	}
	return nil
}

func (art *Tree) Seek(partial Key, userCallback Callback, opts ...int) {
	options := traverseOptions(opts...)
	callback := traverseFilter(options, userCallback)
	target := art.seekPrefix(partial)
	if target != nil {
		art.recursiveTraverse(target, callback)
	}
}
