package parser

import "github.com/bfontaine/quinoa/ast"

type nodeStack struct {
	nodes              []*ast.Node
	length, realLength int
}

func newNodeStack() *nodeStack {
	return &nodeStack{
		nodes: make([]*ast.Node, 0),
	}
}

func (s *nodeStack) Push(n *ast.Node) {
	if s.length < s.realLength {
		s.nodes[s.length] = n
		s.length++
	} else {
		s.nodes = append(s.nodes, n)
		s.length++
		s.realLength++
	}
}

func (s *nodeStack) Pop() *ast.Node {
	if s.length <= 0 {
		return nil
	}

	n := s.Peek()
	s.length--
	return n
}

func (s *nodeStack) Peek() *ast.Node {
	if s.length <= 0 {
		return nil
	}

	return s.nodes[s.length-1]
}
