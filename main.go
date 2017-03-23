package main

import (
	"fmt"
	"log"

	"github.com/bfontaine/quinoa/compiler"

	"llvm.org/llvm/bindings/go/llvm"
)

func main() {
	// See https://felixangell.com/blog/an-introduction-to-llvm-in-go
	builder := llvm.NewBuilder()
	mod := llvm.NewModule("helloworld")

	// return int32; takes nothing; not variadic
	main := llvm.FunctionType(llvm.Int32Type(), []llvm.Type{}, false)
	llvm.AddFunction(mod, "main", main)
	block := llvm.AddBasicBlock(mod.NamedFunction("main"), "entry")
	builder.SetInsertPoint(block, block.FirstInstruction())

	a := builder.CreateAlloca(llvm.Int32Type(), "a")
	// type, value, signed?
	builder.CreateStore(llvm.ConstInt(llvm.Int32Type(), 21, false), a)

	b := builder.CreateAlloca(llvm.Int32Type(), "b")
	builder.CreateStore(llvm.ConstInt(llvm.Int32Type(), 21, false), b)

	aVal := builder.CreateLoad(a, "a_val")
	bVal := builder.CreateLoad(b, "b_val")

	// TODO see how to print that
	// https://stackoverflow.com/questions/31092531/llvm-ir-printing-a-number

	res := builder.CreateAdd(aVal, bVal, "ab_val")
	builder.CreateRet(res)

	comp := compiler.NewCompiler()

	if err := comp.WriteObjectFile(mod, "a.o"); err != nil {
		log.Fatal(err)
	}

	if err := comp.LinkObjectFile("a.o", "a.out", []string{"-lSystem"}); err != nil {
		log.Fatal(err)
	}
	fmt.Println("--> a.out")
}
