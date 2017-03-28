package compiler

import (
	"go/ast"

	"llvm.org/llvm/bindings/go/llvm"
)

var charType = llvm.Int8Type()

func constString(s string) llvm.Value {
	chars := make([]llvm.Value, len(s)+1)
	i := 0

	for _, c := range []rune(s) {
		chars[i] = llvm.ConstInt(charType, uint64(c), false)
		i += 1
	}
	chars[i] = llvm.ConstInt(llvm.Int8Type(), 0, false)

	return llvm.ConstArray(charType, chars)
}

func (c *Compiler) CompileToIR(ast *ast.Node) (*IR, error) {
	// dummy compiler; we just use hardcoded stuff below

	// See https://felixangell.com/blog/an-introduction-to-llvm-in-go
	builder := llvm.NewBuilder()
	mod := llvm.NewModule("helloworld")

	stringType := llvm.PointerType(charType, 0)

	// printf
	printfType := llvm.FunctionType(llvm.Int32Type(), []llvm.Type{stringType}, true)
	//printf.SetFunctionCallConv(llvm.CCallConv)
	printf := llvm.AddFunction(mod, "printf", printfType)
	printf.SetLinkage(llvm.ExternalLinkage)

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

	// format string
	//format := constString("a+b = %d\n")
	format := builder.CreateGlobalString("%d + %d = %d\n", "format")

	formatB := llvm.ConstBitCast(format, llvm.PointerType(charType, 0))

	builder.CreateCall(printf, []llvm.Value{formatB, aVal, bVal, res}, "printf")

	builder.CreateRet(llvm.ConstInt(llvm.Int32Type(), 0, false))

	return &IR{mod: mod}, nil
}
