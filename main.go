package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/bfontaine/quinoa/compiler"
	"github.com/bfontaine/quinoa/parser"
	"github.com/bfontaine/quinoa/vm"
)

func main() {
	var output, ldflags string
	var debug, vmFlag bool

	flag.BoolVar(&debug, "debug", false, "debug")

	flag.StringVar(&output, "o", "a.out", "output file")
	flag.StringVar(&ldflags, "ldflags", "-lc", "comma-separated ld flags")

	flag.BoolVar(&vmFlag, "vm", false, "Run the code in a VM")

	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Please give me one source file")
		os.Exit(1)
	}

	code, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	if debug {
		log.Println("Parsing...")
	}

	ast, err := parser.Parse(string(code), debug)
	if err != nil {
		log.Fatal(err)
	}

	if debug {
		log.Printf("Parsed:\n%v", ast)
	}

	gs, err := compiler.CompileGrains(ast)
	if err != nil {
		log.Fatal(err)
	}

	if vmFlag {
		vm := vm.NewVM(debug)
		if err := vm.Run(gs); err != nil {
			log.Fatal(err)
		}
	}

	//	comp := compiler.NewCompiler()
	//
	//	log.Println("Compiling to IR...")
	//
	//	ir, err := comp.CompileToIR(ast)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	log.Println("Compiling...")
	//
	//	if err := comp.WriteObjectFile(ir, "a.o"); err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	// macOS min: -lSystem
	//	ldflagsArgs := strings.Split(ldflags, ",")
	//
	//	log.Println("Linking...")
	//
	//	if err := comp.LinkObjectFile("a.o", output, ldflagsArgs); err != nil {
	//		log.Fatal(err)
	//	}
}
