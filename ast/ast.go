package ast

import "bytes"

type NodeType int8

const (
	RootNodeType NodeType = iota
	AssignNodeType
	LitteralNodeType
	VariableNodeType
	UnopNodeType
	BinopNodeType
	FuncCallNodeType
)

type Node struct {
	children []*Node
	name     string
	nodeType NodeType
}

func NewNode(nodeType NodeType, name string) *Node {
	return &Node{
		children: make([]*Node, 0),
		name:     name,
		nodeType: nodeType,
	}
}

func (n *Node) Type() NodeType { return n.nodeType }
func (n *Node) Name() string   { return n.name }

func (n1 *Node) AddChild(n2 *Node) {
	n1.children = append(n1.children, n2)
}

func (n *Node) Children() []*Node {
	return n.children
}

func (n *Node) child(i int) *Node {
	if i >= len(n.children) || i < 0 {
		return nil
	}
	return n.children[i]
}

func (n *Node) Child() *Node       { return n.child(0) }
func (n *Node) SecondChild() *Node { return n.child(1) }

func (n *Node) String() string {
	var prefix string

	useName := true

	switch n.nodeType {
	case RootNodeType:
		prefix = "root"
		useName = false
	case AssignNodeType:
		prefix = "assignment"
		useName = false
	case LitteralNodeType:
		prefix = "litteral"
	case VariableNodeType:
		prefix = "var"
	case UnopNodeType:
		prefix = "unop"
	case BinopNodeType:
		prefix = "binop"
	case FuncCallNodeType:
		prefix = "funccall"
	default:
		prefix = "?"
	}

	var b bytes.Buffer
	b.Write([]byte(prefix))
	b.Write([]byte("("))
	if useName {
		b.Write([]byte(n.name))
	}
	for i, ch := range n.Children() {
		if i > 0 || useName {
			b.Write([]byte(", "))
		}
		b.Write([]byte(ch.String()))
	}
	b.Write([]byte(")"))
	return b.String()
}
