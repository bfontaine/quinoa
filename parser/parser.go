package parser

import "github.com/bfontaine/quinoa/ast"

func NewParser(code string) *Parser {
	p := &Parser{
		root:   ast.NewNode(ast.RootNodeType, ""),
		stack:  newNodeStack(),
		Buffer: code,
	}

	p.Init()
	return p
}

func (p *Parser) AST() *ast.Node {
	return p.root
}

// TODO tests

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

func (p *Parser) AddStatement() {
	// |stmt -> |
	p.root.AddChild(p.pop())
}

func (p *Parser) AddAssign() {
	// |... value variable -> |... assign(variable, value)
	value := p.pop()
	variable := p.pop()

	n := ast.NewNode(ast.AssignNodeType, "")
	n.AddChild(variable)
	n.AddChild(value)
	p.push(n)
}

func (p *Parser) AddFuncCall(name string) {
	// |... -> |... funcCall(name)
	p.newNode(ast.FuncCallNodeType, name)
}

func (p *Parser) AddFuncCallArg() {
	arg := p.pop()
	p.last().AddChild(arg)
}

func (p *Parser) AddLitteral(name string) {
	// |... -> |... litteral
	p.newNode(ast.LitteralNodeType, name)
}

func (p *Parser) AddVariable(name string) {
	// |... -> |... variable
	p.newNode(ast.VariableNodeType, name)
}

func (p *Parser) AddUnop(name string) {
	// |... expr -> |... unop(expr)
	n := ast.NewNode(ast.UnopNodeType, name)
	n.AddChild(p.pop())
	p.push(n)
}

func (p *Parser) AddBinop(name string) {
	// |... expr1 expr2 -> |... binop(expr1, expr2)
	n := ast.NewNode(ast.BinopNodeType, name)
	expr2 := p.pop()
	expr1 := p.pop()
	n.AddChild(expr1)
	n.AddChild(expr2)

	p.push(n)
}
