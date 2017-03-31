package vm

import (
	"fmt"
	"log"

	"github.com/bfontaine/quinoa/language"
)

type Value int64

type VM struct {
	memory map[string]Value
	stack  []Value
	top    uint8

	Debug bool
}

func NewVM(debug bool) *VM {
	return &VM{
		memory: make(map[string]Value),
		stack:  make([]Value, 20), // small stack
		Debug:  debug,
	}
}

func (vm *VM) pop() Value {
	vm.top--
	return vm.stack[vm.top]
}
func (vm *VM) push(v Value) {
	vm.stack[vm.top] = v
	vm.top++
}
func (vm *VM) peek() Value {
	return vm.stack[vm.top-1]
}

func (vm *VM) Run(code language.Grains) error {
	for _, inst := range code {
		if vm.Debug {
			log.Printf("vm.next_inst: %+v\nvm.memory: %+v\n", inst, vm.memory)
		}

		switch inst.OpCode {
		case language.DiscardOpCode:
			vm.top--

		case language.StoreOpCode:
			vm.memory[inst.Name] = vm.peek()

		case language.LoadOpCode:
			vm.push(vm.memory[inst.Name])

		case language.ConstOpCode:
			vm.push(Value(inst.Value))

		case language.AddOpCode:
			switch inst.Name {
			case "+":
				e1 := vm.pop()
				e2 := vm.pop()
				vm.push(e1 + e2)
			}

		case language.CallOpCode:
			args := make([]interface{}, 0, inst.PopN)

			for i := 0; i < inst.PopN; i++ {
				args = append(args, interface{}(vm.pop()))
			}

			switch inst.Name {
			case "print":
				fmt.Println(args...)
			default:
				return fmt.Errorf("Unknown function '%s'", inst.Name)
			}

			vm.push(0)
		}
	}

	return nil
}
