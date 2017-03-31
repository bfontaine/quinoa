package compiler

import (
	"github.com/bfontaine/quinoa/ast"
	"github.com/bfontaine/quinoa/language"
)

func CompileGrains(a *ast.Node) (language.Grains, error) {
	var grains language.Grains

	switch a.Type() {
	case ast.RootNodeType:
		for _, ch := range a.Children() {
			if gs, err := CompileGrains(ch); err != nil {
				return nil, err
			} else {
				grains = append(grains, gs...)
				grains = append(grains, language.Grain{language.DiscardOpCode, "", 0, 1})
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

		grains = append(grains, language.Grain{language.StoreOpCode, variable.Name(), a.Value(), 1})

	case ast.LitteralNodeType:
		grains = append(grains, language.Grain{language.ConstOpCode, a.Name(), a.Value(), 0})

	case ast.VariableNodeType:
		grains = append(grains, language.Grain{language.LoadOpCode, a.Name(), a.Value(), 0})

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

		grains = append(grains, language.Grain{language.AddOpCode, a.Name(), a.Value(), 2})

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
		grains = append(grains, language.Grain{language.CallOpCode, a.Name(), a.Value(), nargs})
	}

	return grains, nil
}
