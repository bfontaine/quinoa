package parser

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseValidPrograms(t *testing.T) {
	var msg string
	var err error

	for _, code := range []string{
		"a=1",
		"a=a",
		"a=a+1",
		"a = a + 1",
		"a = a + 1\n",
		"a = a + 1\n\n\t          \n\n\t # end",
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
	} {
		t.Log(strconv.Quote(code))
		_, err = Parse(code)
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
		_, err = Parse(code)
		assert.NotNil(t, err)
	}
}
