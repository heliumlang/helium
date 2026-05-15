/*
 * Compiler for the stack-based VM
 */

package compiler

import (
	"fmt"
	"math"
	"strings"
)

type OpCode uint8

const (
	/* stack */
	OpNop OpCode = iota
	OpPushNone
	OpPushBool
	OpPushInt
	OpPushFloat
	OpPushStr
	OpPushChar
	OpPop
	OpDup
	OpSwap

	/* variables */
	OpLoadLocal
	OpStoreLocal
	OpLoadGlobal
	OpStoreGlobal
	OpLoadUpvalue
	OpStoreUpvalue
	OpCloseUpvalue

	/* arithmetic */
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpNeg

	/* bitwise */
	OpBitAnd
	OpBitOr
	OpBitXor
	OpBitNot
	OpShl
	OpShr

	/* logic */
	OpNot
	OpAnd
	OpOr

	/* comparison */
	OpEq
	OpNeq
	OpLt
	OpLte
	OpGt
	OpGte

	/* control flow */
	OpJmp
	OpJmpFalse
	OpJmpTrue

	/* functions and closures */
	OpCall
	OpCallMethod
	OpReturn
	OpMakeClosure

	/* collections */
	OpMakeArray
	OpMakeMap
	OpIndex
	OpSetIndex
	OpFieldGet
	OpFieldSet

	/* safety/errors */
	OpRaise
	OpUnwrap
	OpCoalesce
	OpIsNone

	/* misc */
	OpPrint // temporary
)

var opNames = [...]string{
	OpNop: "NOP", OpPushNone: "PUSH_NONE", OpPushBool: "PUSH_BOOL",
	OpPushInt: "PUSH_INT", OpPushFloat: "PUSH_FLOAT", OpPushStr: "PUSH_STR",
	OpPushChar: "PUSH_CHAR", OpPop: "POP", OpDup: "DUP", OpSwap: "SWAP",
	OpLoadLocal: "LOAD_LOCAL", OpStoreLocal: "STORE_LOCAL",
	OpLoadGlobal: "LOAD_GLOBAL", OpStoreGlobal: "STORE_GLOBAL",
	OpLoadUpvalue: "LOAD_UPVALUE", OpStoreUpvalue: "STORE_UPVALUE",
	OpCloseUpvalue: "CLOSE_UPVALUE",
	OpAdd:          "ADD", OpSub: "SUB", OpMul: "MUL", OpDiv: "DIV", OpMod: "MOD", OpNeg: "NEG",
	OpBitAnd: "BIT_AND", OpBitOr: "BIT_OR", OpBitXor: "BIT_XOR",
	OpBitNot: "BIT_NOT", OpShl: "SHL", OpShr: "SHR",
	OpNot: "NOT", OpAnd: "AND", OpOr: "OR",
	OpEq: "EQ", OpNeq: "NEQ", OpLt: "LT", OpLte: "LTE", OpGt: "GT", OpGte: "GTE",
	OpJmp: "JMP", OpJmpFalse: "JMP_FALSE", OpJmpTrue: "JMP_TRUE",
	OpCall: "CALL", OpCallMethod: "CALL_METHOD", OpReturn: "RETURN",
	OpMakeClosure: "MAKE_CLOSURE",
	OpMakeArray:   "MAKE_ARRAY", OpMakeMap: "MAKE_MAP",
	OpIndex: "INDEX", OpSetIndex: "SET_INDEX",
	OpFieldGet: "FIELD_GET", OpFieldSet: "FIELD_SET",
	OpRaise: "RAISE", OpUnwrap: "UNWRAP", OpCoalesce: "COALESCE", OpIsNone: "IS_NONE",
	OpPrint: "PRINT",
}

func (o OpCode) String() string {
	if int(o) < len(opNames) {
		return opNames[o]
	}
	return fmt.Sprintf("UNKNOWN(%d)", o)
}

/* instruction */
type Instr struct {
	Op      OpCode
	Operand int // -1 = no operand
	Line    int // line in the source, for errors
}

func (i Instr) String() string {
	if i.Operand == -1 {
		return i.Op.String()
	}
	return fmt.Sprintf("%-18s %d", i.Op.String(), i.Operand)
}

/* constant pool
 *
 * Holds all literal values referenced
 * by a chunk
 *
 * For now it only has 5 arrays with
 * some common types, but this should be
 * changed to support all possible types
 * the language offers.
 */
type ConstPool struct {
	Ints    []int64
	Uints   []uint64
	Floats  []float64
	Strings []string
	Chars   []byte
}

func (p *ConstPool) AddInt(v int64) int {
	for i, x := range p.Ints {
		if x == v {
			return i
		}
	}
	p.Ints = append(p.Ints, v)
	return len(p.Ints) - 1
}

func (p *ConstPool) AddUint(v uint64) int {
	for i, x := range p.Uints {
		if x == v {
			return i
		}
	}
	p.Uints = append(p.Uints, v)
	return len(p.Uints) - 1
}

func (p *ConstPool) AddFloat(v float64) int {
	for i, x := range p.Floats {

		if math.Float64bits(x) == math.Float64bits(v) {
			return i
		}
	}
	p.Floats = append(p.Floats, v)
	return len(p.Floats) - 1
}

func (p *ConstPool) AddString(v string) int {
	for i, x := range p.Strings {
		if x == v {
			return i
		}
	}
	p.Strings = append(p.Strings, v)
	return len(p.Strings) - 1
}

func (p *ConstPool) AddChar(v byte) int {
	for i, x := range p.Chars {
		if x == v {
			return i
		}
	}
	p.Chars = append(p.Chars, v)
	return len(p.Chars) - 1
}

/* upvalue
 *
 * Describes how a closure captures a variable from an
 * enclosing scope.
 */
type Upvalue struct {
	// true -> capture from enclosing function's local stack slot at Index
	// false -> capture from enclosing function's own upvalue list at Index
	IsLocal bool
	Index   int
	Name    string // for dissasembly
}

func (u Upvalue) String() string {
	src := "upvalue"
	if u.IsLocal {
		src = "local"
	}
	return fmt.Sprintf("%s[%d] (%s)", src, u.Index, u.Name)
}

/* local
 *
 * Represents a local variable.
 */
type Local struct {
	Name  string
	Depth int
	Slot  int
	Const bool
}

/* chunk
 *
 * Chunk is the unit of compilation, every functions (including top-level) compiles
 * to one chunk.
 */
type Chunk struct {
	Name     string
	Arity    int // number of args
	Code     []Instr
	Pool     ConstPool
	Locals   []Local
	Upvalues []Upvalue
}

// appends an instruction and returns the index
func (c *Chunk) Emit(op OpCode, operand, line int) int {
	c.Code = append(c.Code, Instr{Op: op, Operand: operand, Line: line})
	return len(c.Code) - 1
}

// shorthand for instructions with no operand
func (c *Chunk) EmitOp(op OpCode, line int) int {
	return c.Emit(op, -1, line)
}

// emits a jump with a dummy target, patch once the target is known
func (c *Chunk) Placeholder(op OpCode, line int) int {
	return c.Emit(op, -1, line)
}

// patch writes the current EOC (end-of-code) position into a
// previously emitted jump instruction
func (c *Chunk) Patch(instrIdx int) {
	c.Code[instrIdx].Operand = len(c.Code)
}

// writes an explicit target into a jump instruction
func (c *Chunk) PatchTo(instrIdx, target int) {
	c.Code[instrIdx].Operand = target
}

func (c *Chunk) Len() int { return len(c.Code) }

func (c *Chunk) AddUpvalue(name string, isLocal bool, index int) int {

	for i, u := range c.Upvalues {
		if u.IsLocal == isLocal && u.Index == index {
			return i
		}
	}
	c.Upvalues = append(c.Upvalues, Upvalue{IsLocal: isLocal, Index: index, Name: name})
	return len(c.Upvalues) - 1
}

// for debugging, do not use in the real world
func (c *Chunk) Disassemble() string {
	var b strings.Builder
	fmt.Fprintf(&b, "=== %s (arity=%d) ===\n", c.Name, c.Arity)

	if len(c.Pool.Ints) > 0 {
		fmt.Fprintf(&b, "  ints:    %v\n", c.Pool.Ints)
	}
	if len(c.Pool.Floats) > 0 {
		fmt.Fprintf(&b, "  floats:  %v\n", c.Pool.Floats)
	}
	if len(c.Pool.Strings) > 0 {
		fmt.Fprintf(&b, "  strings: %v\n", c.Pool.Strings)
	}
	if len(c.Upvalues) > 0 {
		fmt.Fprintf(&b, "  upvalues:\n")
		for i, u := range c.Upvalues {
			fmt.Fprintf(&b, "    [%d] %s\n", i, u)
		}
	}
	b.WriteString("\n")

	for i, instr := range c.Code {

		annotation := ""
		switch instr.Op {
		case OpPushInt:
			if instr.Operand >= 0 && instr.Operand < len(c.Pool.Ints) {
				annotation = fmt.Sprintf("; %d", c.Pool.Ints[instr.Operand])
			}
		case OpPushFloat:
			if instr.Operand >= 0 && instr.Operand < len(c.Pool.Floats) {
				annotation = fmt.Sprintf("; %g", c.Pool.Floats[instr.Operand])
			}
		case OpPushStr:
			if instr.Operand >= 0 && instr.Operand < len(c.Pool.Strings) {
				annotation = fmt.Sprintf("; %q", c.Pool.Strings[instr.Operand])
			}
		case OpLoadLocal, OpStoreLocal:
			if instr.Operand >= 0 && instr.Operand < len(c.Locals) {
				annotation = "; " + c.Locals[instr.Operand].Name
			}
		case OpLoadUpvalue, OpStoreUpvalue:
			if instr.Operand >= 0 && instr.Operand < len(c.Upvalues) {
				annotation = "; " + c.Upvalues[instr.Operand].Name
			}
		case OpJmp, OpJmpFalse, OpJmpTrue:
			annotation = fmt.Sprintf("; → %d", instr.Operand)
		}

		fmt.Fprintf(&b, "  %04d  line %-4d  %s%s\n",
			i, instr.Line, instr.String(), annotation)
	}
	return b.String()
}

/* scope
 *
 * Tracks locals and current nesting depth during compilation.
 *
 * Each function has one scope, and they do nost nest across functions.
 */
type Scope struct {
	Locals []Local
	Depth  int // 0 = top-level
}

// increase depth
func (s *Scope) Begin() { s.Depth++ }

// closes the current block and returns locals that went
// out of scope
func (s *Scope) End() []Local {
	s.Depth--
	var popped []Local
	for len(s.Locals) > 0 && s.Locals[len(s.Locals)-1].Depth > s.Depth {
		popped = append(popped, s.Locals[len(s.Locals)-1])
		s.Locals = s.Locals[:len(s.Locals)-1]
	}
	return popped
}

// adds a new local and returns its slot index
func (s *Scope) Declare(name string, isConst bool) (int, error) {
	for i := len(s.Locals) - 1; i >= 0; i-- {
		l := s.Locals[i]
		if l.Depth < s.Depth {
			break
		}
		if l.Name == name {
			return -1, fmt.Errorf("variable %q already declared in this scope", name)
		}
	}
	slot := len(s.Locals)
	s.Locals = append(s.Locals, Local{
		Name:  name,
		Depth: s.Depth,
		Slot:  slot,
		Const: isConst,
	})
	return slot, nil
}

// walks backwards to find a local by-name
func (s *Scope) Resolve(name string) (Local, bool) {
	for i := len(s.Locals) - 1; i >= 0; i-- {
		if s.Locals[i].Name == name {
			return s.Locals[i], true
		}
	}
	return Local{}, false
}

/* loop context
 *
 * Tracks jump targets that need to be patched when a loop ends.
 */
type LoopCtx struct {
	Start      int
	BreakJumps []int
}

// patch all break jumps
func (l *LoopCtx) PatchBreaks(chunk *Chunk) {
	for _, idx := range l.BreakJumps {
		chunk.Patch(idx)
	}
	l.BreakJumps = nil
}

/* function context
 *
 * Holds all mutable state for compiling a single function.
 * One per FunctionDecl or ClosureExpr.
 */
type FnCtx struct {
	Chunk     *Chunk
	Scope     Scope
	Loops     []LoopCtx // stack; innermost loop is last
	Enclosing *FnCtx    // nil for top-level
}

func NewFnCtx(name string, arity int, enclosing *FnCtx) *FnCtx {
	return &FnCtx{
		Chunk:     &Chunk{Name: name, Arity: arity},
		Enclosing: enclosing,
	}
}

// calls f.Chunk.Emit
func (f *FnCtx) Emit(op OpCode, operand, line int) int {
	return f.Chunk.Emit(op, operand, line)
}

// calls f.Chunk.EmitOp
func (f *FnCtx) EmitOp(op OpCode, line int) int {
	return f.Chunk.EmitOp(op, line)
}

// create a loop ctx
func (f *FnCtx) PushLoop(start int) {
	f.Loops = append(f.Loops, LoopCtx{Start: start})
}

// closes the innermost loop ctx and patches its
// break jumps
func (f *FnCtx) PopLoop() {
	if len(f.Loops) == 0 {
		return
	}
	top := &f.Loops[len(f.Loops)-1]
	top.PatchBreaks(f.Chunk)
	f.Loops = f.Loops[:len(f.Loops)-1]
}

// pointer to the innermost loop
func (f *FnCtx) CurrentLoop() *LoopCtx {
	if len(f.Loops) == 0 {
		return nil
	}
	return &f.Loops[len(f.Loops)-1]
}

// walks the enclosing FnCtx chain to capture a variable,
// returns the upvalue index in f.Chunk, or -1 if not found
func (f *FnCtx) ResolveUpvalue(name string) int {
	if f.Enclosing == nil {
		return -1
	}

	if local, ok := f.Enclosing.Scope.Resolve(name); ok {
		return f.Chunk.AddUpvalue(name, true, local.Slot)
	}

	if uvIdx := f.Enclosing.ResolveUpvalue(name); uvIdx != -1 {
		return f.Chunk.AddUpvalue(name, false, uvIdx)
	}
	return -1
}
