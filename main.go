package main

import (
	"fmt"
	"log"

	"github.com/bfontaine/quinoa/compiler"
	"github.com/bfontaine/quinoa/language"
)

func main() {
	ast, err := language.Parse("")
	if err != nil {
		log.Fatal(err)
	}

	comp := compiler.NewCompiler()

	ir, err := comp.CompileToIR(ast)
	if err != nil {
		log.Fatal(err)
	}

	if err := comp.WriteObjectFile(ir, "a.o"); err != nil {
		log.Fatal(err)
	}

	if err := comp.LinkObjectFile("a.o", "a.out", []string{"-lSystem", "-lc"}); err != nil {
		log.Fatal(err)
	}
	fmt.Println("--> a.out")
}
