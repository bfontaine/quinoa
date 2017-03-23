package compiler

import (
	"os"
	"os/exec"

	"llvm.org/llvm/bindings/go/llvm"
)

type Compiler struct{}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) WriteObjectFile(mod llvm.Module, filename string) error {

	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		return err
	}

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
		return err
	}

	tm := t.CreateTargetMachine(tt, "", "", llvm.CodeGenLevelNone, llvm.RelocDefault, llvm.CodeModelDefault)

	mod.SetDataLayout("")
	mod.SetTarget(tt)

	buff, err := tm.EmitToMemoryBuffer(mod, llvm.ObjectFile)
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write(buff.Bytes())

	return nil
}

func (c *Compiler) LinkObjectFile(source, target string, flags []string) error {
	args := append([]string{source, "-o", target}, flags...)

	cmd := exec.Command("ld", args...)
	return cmd.Run()
}
