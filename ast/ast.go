package ast

type NodeType int8

const (
	RootNodeType NodeType = iota
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

func (n1 *Node) AddChild(n2 *Node) {
	n1.children = append(n1.children, n2)
}

func (n *Node) Children() []*Node {
	return n.children
}

func (n *Node) String() string {
	return n.name
}
