package ast

// node is the struct implements node interface except for Accept method.
// Node implementations should embed it in.
type node struct {
	text string
}

// SetText implements Node interface.
func (n *node) SetText(text string) {
	n.text = text
}

// Text implements Node interface.
func (n *node) Text() string {
	return n.text
}

// Node is the basic element of the AST.
// Interfaces embed Node should have 'Node' name suffix.
type Node interface {
	// Text returns the original text of the element.
	Text() string
	// SetText sets original text to the Node.
	SetText(text string)
}

// ExprNode is a expr that can be evaluated.
// Name of implementations should have 'Expr' suffix.
type ExprNode interface {
	// Node is embedded in ExprNode.
	Node
	// SetType sets evaluation type to the expression.
	SetType(tp uint8)
	// GetType gets the evaluation type of the expression.
	GetType() uint8
}
