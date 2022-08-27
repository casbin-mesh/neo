package iterator

import "github.com/casbin-mesh/neo/pkg/expression/ast"

type BfsIterator interface {
	Iterator
}

type bfsIterator struct {
	queue []Stack
}

func (b *bfsIterator) Next() ast.Evaluable {
	if !b.HasNext() {
		return nil
	}
	var head *Stack

	head = &b.queue[0]
	curNode := head.node
	curIndex := head.index

	if curIndex < curNode.ChildrenLen() {
		for i := curIndex; i < curNode.ChildrenLen(); i++ {
			nextNode := curNode.GetChildAt(i)
			if nextNode != nil {
				b.queue = append(b.queue, Stack{node: nextNode, index: 0})
			}
		}
		head.index = curNode.ChildrenLen()
	}

	b.queue = b.queue[1:]
	return curNode
}

func (b *bfsIterator) nextChild(cur ast.Evaluable, idx int) (int, ast.Evaluable) {
	if idx == -1 {
		idx = 0
	}
	for i := idx; i < cur.ChildrenLen(); i++ {
		child := cur.GetChildAt(i)
		if child != nil {
			return i + 1, child
		}
	}
	return 0, nil
}

func (b *bfsIterator) HasNext() bool {
	return len(b.queue) != 0
}

func NewBfsIterator(root ast.Evaluable) BfsIterator {
	return &bfsIterator{
		queue: []Stack{{node: root, index: -1}},
	}
}
