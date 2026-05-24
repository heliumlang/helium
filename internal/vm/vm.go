package vm

import (
	"fmt"

	"github.com/heliumlang/helium/internal/compiler"
	"github.com/heliumlang/helium/internal/util"
)

type Frame struct {
	fn     compiler.FuncMeta // the function being executed
	ip     int               // instruction pointer
	locals []any             // local variables for this call
}

type Native struct {
	name  string
	arity int
	fn    func(args []any) []any
}

type VM struct {
	constants []compiler.Const
	functions []compiler.FuncMeta
	natives   []Native
	stack     *util.Stack[any]
	frames    []Frame
}

type StandardLibrary map[string]func(args []any) []any

func New(program compiler.Program, std StandardLibrary) *VM {
	return &VM{
		constants: program.Constants,
		functions: program.Functions,
		natives:   mapNatives(program.Natives, std),
		stack:     util.NewStack[any](),
	}
}

func mapNatives(compileNatives []compiler.Native, std StandardLibrary) []Native {
	var natives []Native
	for _, v := range compileNatives {
		fn, ok := std[v.Name()]
		if !ok {
			continue
		}
		natives = append(natives, Native{
			name:  v.Name(),
			arity: v.ArgCount(),
			fn:    fn,
		})
	}
	return natives
}

// Thin wrapper of VM.stack.Push()
func (vm *VM) push(v any) int {
	return vm.stack.Push(v)
}

// Thin wrapper of VM.stack.Pop()
func (vm *VM) pop() any {
	return vm.stack.Pop()
}

// Thin wrapper of VM.stack.Get()
func (vm *VM) get(i int) (any, error) {
	return vm.stack.Get(i)
}

// Thin wrapper of VM.stack.Slice()
func (vm *VM) wholeStack() []any {
	return vm.stack.Slice()
}

// Run the program.
func (vm *VM) Run() error {
	for i, fn := range vm.functions {
		if fn.Name() == "main" {
			vm.pushFrame(i)
			return vm.execute()
		}
	}
	return fmt.Errorf("no main function found")
}

// Push a new frame.
func (vm *VM) pushFrame(fnIdx int) {
	fn := vm.functions[fnIdx]
	frame := Frame{
		fn:     fn,
		ip:     0,
		locals: make([]any, fn.ArgCount()+fn.LocalCount()),
	}
	vm.frames = append(vm.frames, frame)
}

// Pop a frame.
func (vm *VM) popFrame() Frame {
	frame := vm.frames[len(vm.frames)-1]
	vm.frames = vm.frames[:len(vm.frames)-1]
	return frame
}

// Get current frame.
func (vm *VM) frame() *Frame {
	return &vm.frames[len(vm.frames)-1]
}

// Execute operations.
func (vm *VM) execute() error {
	for {
		frame := vm.frame()
		op, operands := vm.decode(frame)

		switch op {
		case compiler.OpNoop:

		case compiler.OpPushConst:
			vm.push(vm.constants[operands[0]].Value())

		case compiler.OpPushNone:
			vm.push(nil)

		case compiler.OpPushArr:
			count := int(operands[0])
			elems := make([]any, count)
			for i := count - 1; i >= 0; i-- {
				elems[i] = vm.pop()
			}
			vm.push(elems)

		case compiler.OpPushMap:
			count := int(operands[0])
			m := make(map[any]any, count)
			for range count {
				val := vm.pop()
				key := vm.pop()
				m[key] = val
			}
			vm.push(m)

		case compiler.OpStoreLocal:
			frame.locals[operands[0]] = vm.pop()

		case compiler.OpLoadLocal:
			vm.push(frame.locals[operands[0]])

		case compiler.OpAddInt:
			b, a := vm.pop().(int64), vm.pop().(int64)
			vm.push(a + b)

		case compiler.OpSubInt:
			b, a := vm.pop().(int64), vm.pop().(int64)
			vm.push(a - b)

		case compiler.OpMulInt:
			b, a := vm.pop().(int64), vm.pop().(int64)
			vm.push(a * b)

		case compiler.OpDivInt:
			b, a := vm.pop().(int64), vm.pop().(int64)
			if b == 0 {
				return fmt.Errorf("division by zero")
			}
			vm.push(a / b)

		case compiler.OpModInt:
			b, a := vm.pop().(int64), vm.pop().(int64)
			vm.push(a % b)

		case compiler.OpNegInt:
			vm.push(-vm.pop().(int64))

		case compiler.OpAddFloat:
			b, a := vm.pop().(float64), vm.pop().(float64)
			vm.push(a + b)

		case compiler.OpSubFloat:
			b, a := vm.pop().(float64), vm.pop().(float64)
			vm.push(a - b)

		case compiler.OpMulFloat:
			b, a := vm.pop().(float64), vm.pop().(float64)
			vm.push(a * b)

		case compiler.OpDivFloat:
			b, a := vm.pop().(float64), vm.pop().(float64)
			if b == 0 {
				return fmt.Errorf("division by zero")
			}
			vm.push(a / b)

		case compiler.OpNegFloat:
			vm.push(-vm.pop().(float64))

		case compiler.OpEq:
			vm.push(vm.pop() == vm.pop())

		case compiler.OpNeq:
			vm.push(vm.pop() != vm.pop())

		case compiler.OpLt:
			b, a := vm.pop().(int64), vm.pop().(int64)
			vm.push(a < b)

		case compiler.OpGt:
			b, a := vm.pop().(int64), vm.pop().(int64)
			vm.push(a > b)

		case compiler.OpLte:
			b, a := vm.pop().(int64), vm.pop().(int64)
			vm.push(a <= b)

		case compiler.OpGte:
			b, a := vm.pop().(int64), vm.pop().(int64)
			vm.push(a >= b)

		case compiler.OpAnd:
			b, a := vm.pop().(bool), vm.pop().(bool)
			vm.push(a && b)

		case compiler.OpOr:
			b, a := vm.pop().(bool), vm.pop().(bool)
			vm.push(a || b)

		case compiler.OpNot:
			vm.push(!vm.pop().(bool))

		case compiler.OpInc:
			frame.locals[operands[0]] = frame.locals[operands[0]].(int64) + 1

		case compiler.OpDec:
			frame.locals[operands[0]] = frame.locals[operands[0]].(int64) - 1

		case compiler.OpJmp:
			frame.ip = int(operands[0])

		case compiler.OpJmpIfTrue:
			if vm.pop().(bool) {
				frame.ip = int(operands[0])
			}

		case compiler.OpJmpIfFalse:
			if !vm.pop().(bool) {
				frame.ip = int(operands[0])
			}

		case compiler.OpCall:
			fnIdx := int(operands[0])
			argCount := int(operands[1])
			args := make([]any, argCount)
			for i := argCount - 1; i >= 0; i-- {
				args[i] = vm.pop()
			}
			vm.pushFrame(fnIdx)
			for i, arg := range args {
				vm.frame().locals[i] = arg
			}

		case compiler.OpCallNative:
			nativeIdx := int(operands[0])
			argCount := int(operands[1])
			args := make([]any, argCount)
			for i := argCount - 1; i >= 0; i-- {
				args[i] = vm.pop()
			}
			result := vm.natives[nativeIdx].fn(args)
			if result != nil {
				vm.push(result)
			}

		case compiler.OpReturn:
			count := int(operands[0])
			results := make([]any, count)
			for i := count - 1; i >= 0; i-- {
				results[i] = vm.pop()
			}
			vm.popFrame()
			if len(vm.frames) == 0 {
				return nil
			}
			for _, r := range results {
				vm.push(r)
			}

		// -- iterators (implement when you add them to VM) --
		// case compiler.OpMakeIter: ...
		// case compiler.OpNextIter: ...

		default:
			return fmt.Errorf("unknown opcode: %s", op)
		}
	}
}

// Decode next operation.
func (vm *VM) decode(frame *Frame) (compiler.OpCode, []compiler.ByteCode) {
	b := frame.fn.Body().Code()[frame.ip]
	op, count := b.Decode()
	frame.ip++
	operands := frame.fn.Body().ReadOperands(frame.ip, count)
	frame.ip += int(count)
	return op, operands
}
