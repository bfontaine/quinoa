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

func (p *Parser) AddRoot() {
	// | -> |root
	p.newNode(ast.RootNodeType, "")
}

func (p *Parser) AddStatement() {
	// |root stmt -> root(..., stmt)
	stmt := p.pop()
	p.last().AddChild(stmt)
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

func (p *Parser) AddFuncName(name string) {
	// |... -> |... funcName

	// TODO use a token name instead of abusing the AST here
	p.newNode(ast.FuncNameNodeType, name)
}

func (p *Parser) AddFuncCall() {
	// |... fn arg1 ... argn -> |... funcCall(fn, arg1, ..., argn)

	args := make([]*ast.Node, 0)

	for !p.stack.Empty() {
		arg := p.stack.Peek()
		if arg.Type() == ast.FuncNameNodeType {
			p.pop()
			n := ast.NewNode(ast.FuncCallNodeType, arg.Name())
			for i := len(args) - 1; i >= 0; i-- {
				n.AddChild(args[i])
			}
			p.push(n)
			break
		}
		args = append(args, p.stack.Pop())
	}
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
