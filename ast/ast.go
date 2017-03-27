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
	FuncNameNodeType
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
	var useName bool
	var prefix string

	switch n.nodeType {
	case RootNodeType:
		prefix = "root"
	case AssignNodeType:
		prefix = "assignment"
	case LitteralNodeType:
		prefix = "litteral"
		useName = true
	case VariableNodeType:
		prefix = "var"
		useName = true
	case UnopNodeType:
		prefix = "unop"
		useName = true
	case BinopNodeType:
		prefix = "binop"
		useName = true
	case FuncCallNodeType:
		prefix = "funccall"
		useName = true
	default:
		prefix = "?"
		useName = true
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
