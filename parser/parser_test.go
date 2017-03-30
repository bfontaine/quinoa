package parser

import (
	"strconv"
	"testing"

	"github.com/bfontaine/quinoa/ast"
	"github.com/stretchr/testify/assert"
)

type dummyAST struct {
	name     string
	ntype    ast.NodeType
	children []dummyAST
}

func assertEqualASTs(t *testing.T, expected dummyAST, actual *ast.Node) {
	assert.NotNil(t, actual)

	assert.Equal(t, expected.name, actual.Name())
	assert.Equal(t, expected.ntype, actual.Type())

	actualChildren := actual.Children()

	if expected.children == nil {
		assert.Len(t, actualChildren, 0)
		return
	}

	assert.Len(t, actualChildren, len(expected.children))

	for i := 0; i < len(expected.children); i++ {
		assertEqualASTs(t, expected.children[i], actualChildren[i])
	}
}

func TestParseValidPrograms(t *testing.T) {
	var msg string
	var err error

	for _, code := range []string{
		"a=1",
		"a=1\n",
		"a=1#\n",
		"a=1#x\n",
		"a=a",
		"a=a\n",
		"a=a+1",
		"a = a + 1",
		"a = a + 1\n",
		"a = a + 1\n\n\t          \n\n\t # end\n",
		"print(a + a)",
		"print(a, a, a, a, a, a)",
		"print(a, # hey\n b)",
		"print(a # hey\n, b)",
		"f()",
		"f    (\n\n1\n\n)",
		"f(g())",
		"f(g(1))",
		"f(g(h(), i(), j()), k(42))",
		"a=f()",
		"a = f()",
		"a=1\na=2\na=3\n",
		"f(\n\t1,\n\t2,\n)",
	} {
		t.Log(strconv.Quote(code))
		_, err = Parse(code, false)
		if err != nil {
			msg = err.Error()
		} else {
			msg = ""
		}
		assert.Nil(t, err, msg)
	}
}

func TestParseInvalidPrograms(t *testing.T) {
	var err error

	for _, code := range []string{
		"a=",
		"a = 1 + ",
		"(((((((((((",
		"print(a",
		"++",
		"# empty program",
		"f\n\n(\n\n1\n\n)",
	} {
		t.Log(strconv.Quote(code))
		_, err = Parse(code, false)
		assert.NotNil(t, err)
	}
}

func TestParseASTAssignBinopPlus(t *testing.T) {
	actualAST, err := Parse("i = 1 + 2", true)
	assert.Nil(t, err)

	assertEqualASTs(t, dummyAST{"", ast.RootNodeType, []dummyAST{
		dummyAST{"", ast.AssignNodeType, []dummyAST{
			dummyAST{"i", ast.VariableNodeType, nil},
			dummyAST{"+", ast.BinopNodeType, []dummyAST{
				dummyAST{"1", ast.LitteralNodeType, nil},
				dummyAST{"2", ast.LitteralNodeType, nil},
			}},
		}},
	}}, actualAST)
}
