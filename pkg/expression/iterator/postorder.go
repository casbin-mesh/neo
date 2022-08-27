package iterator

import "github.com/casbin-mesh/neo/pkg/expression/ast"

type PostOrderIterator interface {
	Iterator
	NextWithMutParent() (next ast.Evaluable, parent *ast.Evaluable, childIdx int)
}

type postorderStack struct {
	parent *ast.Evaluable
	node   *ast.Evaluable
	index  int
}

type postorderIterator struct {
	root   *ast.Evaluable
	filter Filter
	first  []postorderStack
	second []postorderStack
}

func (p *postorderIterator) Next() ast.Evaluable {
	next, _, _ := p.NextWithMutParent()
	return next
}

func (p *postorderIterator) HasNext() bool {
	return len(p.second) != 0
}

func (p *postorderIterator) NextWithMutParent() (next ast.Evaluable, parent *ast.Evaluable, idx int) {
	var (
		stack postorderStack
		l     int
	)
	for len(p.first) > 0 {
		l = len(p.first)
		stack, p.first = p.first[l-1], p.first[:l-1]

		if stack.node != nil && *stack.node != nil && p.filter(*stack.node) {
			p.second = append(p.second, stack)

			for i := 0; i < (*stack.node).ChildrenLen(); i++ {

				child := (*stack.node).GetMutChildAt(i)

				if child != nil && *child != nil && p.filter(*child) {
					p.first = append(p.first, postorderStack{
						parent: stack.node,
						node:   child,
						index:  i,
					})
				}
			}
		}
		// continue
	}
	if p.HasNext() {
		l = len(p.second)
		stack, p.second = p.second[l-1], p.second[:l-1]

		return *stack.node, stack.parent, stack.index
	}
	return nil, nil, 0
}

func NewPostOrderIterator(root *ast.Evaluable, opts ...*Option) PostOrderIterator {
	opt := getOpt(opts...)
	return &postorderIterator{
		root:   root,
		filter: opt.Filter,
		first:  []postorderStack{{node: root}},
	}
}
