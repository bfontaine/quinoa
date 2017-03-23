package main

import (
	"log"
	"os"

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

	res := builder.CreateAdd(aVal, bVal, "ab_val")
	builder.CreateRet(res)

	if ok := llvm.VerifyModule(mod, llvm.ReturnStatusAction); ok != nil {
		log.Fatal(ok.Error())
	}

	mod.Dump()

	// http://llvm.org/docs/tutorial/LangImpl08.html
	const cpu = "generic"
	const features = ""

	llvm.InitializeAllTargetInfos()
	llvm.InitializeAllTargetMCs()
	llvm.InitializeAllTargets()
	llvm.InitializeAllAsmParsers()
	llvm.InitializeAllAsmPrinters()

	tt := llvm.DefaultTargetTriple()

	t, err := llvm.GetTargetFromTriple(tt)
	if err != nil {
		log.Fatal(err)
	}

	tm := t.CreateTargetMachine(tt, "", "", llvm.CodeGenLevelNone, llvm.RelocDefault, llvm.CodeModelDefault)

	mod.SetDataLayout("")
	mod.SetTarget(tt)

	buff, err := tm.EmitToMemoryBuffer(mod, llvm.ObjectFile)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("a.o")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.Write(buff.Bytes())
}
