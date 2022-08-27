package iterator

import "github.com/casbin-mesh/neo/pkg/expression/ast"

type DfsIterator interface {
	Iterator
}

func NewDfsIterator(root ast.Evaluable, opts ...*Option) DfsIterator {
	opt := getOpt(opts...)
	return &dfsIterator{
		nextNode: root,
		stack:    []Stack{{node: root, index: -1}},
		filter:   opt.Filter,
	}
}

type dfsIterator struct {
	nextNode ast.Evaluable
	cursor   int
	stack    []Stack
	filter   Filter
}

func (b *dfsIterator) Next() ast.Evaluable {
	if !b.HasNext() {
		return nil
	}
	cur := b.nextNode
	b.next()
	return cur
}

func (b *dfsIterator) next() {
	var nextNode ast.Evaluable
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
				newStack := make([]Stack, b.cursor+10)
				copy(newStack, b.stack)
				b.stack = newStack
			}
			b.cursor++
			b.stack[b.cursor] = Stack{node: nextNode, index: -1}
			return
		}
	}
}

func (b *dfsIterator) nextChild(cur ast.Evaluable, idx int) (int, ast.Evaluable) {
	if idx == -1 {
		idx = 0
	}
	for i := idx; i < cur.ChildrenLen(); i++ {
		child := cur.GetChildAt(i)
		if child != nil {
			if b.filter == nil || (b.filter != nil && b.filter(child)) {
				return i + 1, child
			}
		}
	}
	return 0, nil
}

func (b *dfsIterator) HasNext() bool {
	return b.nextNode != nil
}
