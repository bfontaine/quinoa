package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/bfontaine/quinoa/parser"
)

func main() {
	var output, ldflags string

	flag.StringVar(&output, "o", "a.out", "output file")
	flag.StringVar(&ldflags, "ldflags", "-lc", "comma-separated ld flags")

	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Please give me one source file")
		os.Exit(1)
	}

	code, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Parsing...")

	ast, err := parser.Parse(string(code))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Parsed:\n%v", ast)

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
