package compiler

import "llvm.org/llvm/bindings/go/llvm"

type IR struct {
	mod llvm.Module
}
