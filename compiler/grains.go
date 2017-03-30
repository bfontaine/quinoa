package compiler

import "github.com/bfontaine/quinoa/ast"

type OpCode int8

// load(name) -- push 1
// const(value) -- push 1
// add() -- pop 2, push 1
// store(name) -- peek 1
// call(name, N) -- pop N, push 1
// discard() -- pop 1

const (
	StoreOpCode OpCode = iota
	LoadOpCode
	ConstOpCode
	AddOpCode
	CallOpCode
	DiscardOpCode
)

// A Grain represents an instruction in the intermediate representation
type Grain struct {
	OpCode OpCode
	Name   string
	PopN   int
}

// Grains represents a sequence of instructions in the intermediate
// representation.
type Grains []Grain

// TODO tests

func CompileGrains(a *ast.Node) (Grains, error) {
	var grains Grains

	switch a.Type() {
	case ast.RootNodeType:
		for _, ch := range a.Children() {
			if gs, err := CompileGrains(ch); err != nil {
				return nil, err
			} else {
				grains = append(grains, gs...)
				grains = append(grains, Grain{DiscardOpCode, "", 1})
			}
		}

	case ast.AssignNodeType:
		variable := a.Child()
		expr := a.SecondChild()

		if gs, err := CompileGrains(expr); err != nil {
			return nil, err
		} else {
			grains = append(grains, gs...)
		}

		grains = append(grains, Grain{StoreOpCode, variable.Name(), 1})

	case ast.LitteralNodeType:
		grains = append(grains, Grain{ConstOpCode, a.Name(), 0})

	case ast.VariableNodeType:
		grains = append(grains, Grain{LoadOpCode, a.Name(), 0})

	case ast.UnopNodeType:
		if name := a.Name(); name != "+" {
			panic("Unsupported unop: " + name)
		}

		// Discard the +: +1 -> 1
		return CompileGrains(a.Child())

	case ast.BinopNodeType:
		if name := a.Name(); name != "+" {
			panic("Unsupported binop: " + name)
		}

		left := a.Child()
		right := a.SecondChild()

		for _, expr := range []*ast.Node{right, left} {
			if gs, err := CompileGrains(expr); err != nil {
				return nil, err
			} else {
				grains = append(grains, gs...)
			}
		}

		grains = append(grains, Grain{AddOpCode, a.Name(), 2})

	case ast.FuncCallNodeType:
		args := a.Children()
		nargs := len(args)

		for i := nargs - 1; i >= 0; i-- {
			if gs, err := CompileGrains(args[i]); err != nil {
				return nil, err
			} else {
				grains = append(grains, gs...)
			}
		}
		grains = append(grains, Grain{CallOpCode, a.Name(), nargs})
	}

	return grains, nil
}
