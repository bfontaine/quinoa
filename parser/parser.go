package parser

import "github.com/bfontaine/quinoa/ast"

func NewParser(code string) *Parser {
	p := &Parser{
		stack:  newNodeStack(),
		Buffer: code,
	}

	p.AddRoot()

	p.Init()
	return p
}

func (p *Parser) AST() *ast.Node {
	// should be the last node
	return p.stack.Pop()
}

func Parse(code string) (*ast.Node, error) {
	p := NewParser(code)

	if err := p.Parse(); err != nil {
		return nil, err
	}

	p.Execute()

	return p.AST(), nil
}

func (p *Parser) push(n *ast.Node) {
	p.stack.Push(n)
}

func (p *Parser) pop() *ast.Node {
	return p.stack.Pop()
}

func (p *Parser) last() *ast.Node {
	return p.stack.Peek()
}

func (p *Parser) newNode(nodeType ast.NodeType, name string) {
	p.push(ast.NewNode(nodeType, name))
}

func (p *Parser) AddRoot() {
	// | -> |root
	p.newNode(ast.RootNodeType, "")
}

func (p *Parser) AddStatement() {
	// |root stmt -> root
	stmt := p.pop()
	p.last().AddChild(stmt)
}

func (p *Parser) AddAssign()              {}
func (p *Parser) AddFnCall(name string)   {}
func (p *Parser) AddLitteral(name string) {}
func (p *Parser) AddVariable(name string) {}
func (p *Parser) AddUnOp(name string)     {}
func (p *Parser) AddBinOp(name string)    {}
