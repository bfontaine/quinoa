package language

type OpCode int8

// load(name) -- push 1
// const(value) -- push 1
// add() -- pop 2, push 1
// store(name) -- peek 1
// call(name, N) -- pop N, push 1
// discard() -- pop 1

const (
	StoreOpCode OpCode = iota
	LoadOpCode
	ConstOpCode
	AddOpCode
	CallOpCode
	DiscardOpCode
)

// A Grain represents an instruction in the intermediate representation
type Grain struct {
	OpCode OpCode
	Name   string
	Value  int64
	PopN   int
}

// Grains represents a sequence of instructions in the intermediate
// representation.
type Grains []Grain

// TODO tests
