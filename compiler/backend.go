package compiler

import (
	"os"
	"os/exec"

	"llvm.org/llvm/bindings/go/llvm"
)

// CompileIR generates object code from LLVM IR
func (c *Compiler) CompileIR(ir *IR) ([]byte, error) {
	mod := ir.mod

	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		return nil, err
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
		return nil, err
	}

	tm := t.CreateTargetMachine(tt, "", "", llvm.CodeGenLevelNone, llvm.RelocDefault, llvm.CodeModelDefault)

	mod.SetDataLayout("")
	mod.SetTarget(tt)

	mb, err := tm.EmitToMemoryBuffer(mod, llvm.ObjectFile)
	if err != nil {
		return nil, err
	}
	return mb.Bytes(), nil
}

func (c *Compiler) WriteObjectFile(ir *IR, filename string) error {
	buff, err := c.CompileIR(ir)
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write(buff)

	return nil
}

func (c *Compiler) LinkObjectFile(source, target string, flags []string) error {
	args := append([]string{source, "-o", target}, flags...)

	cmd := exec.Command("ld", args...)
	return cmd.Run()
}
