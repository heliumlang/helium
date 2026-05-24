package compiler

import (
	"fmt"
	"strings"

	"github.com/heliumlang/helium/internal/types"
)

type FuncMeta struct {
	name   string
	args   uint16
	locals uint16
	chunk  *Chunk
}

type localVar struct {
	name  string
	depth int
}

type nativeFunc struct {
	name  string
	arity int
}

type extern struct {
	name string
	t    *types.Type
}

type Compiler struct {
	module     string
	current    *Chunk
	constPool  []Const
	functions  []FuncMeta
	natives    []nativeFunc
	locals     []localVar
	externs    []extern
	scopeDepth int
	err        error
	ttable     *types.TypeTable
	inFunc     bool
}

func NewCompiler() *Compiler {
	c := &Compiler{
		module:  "",
		current: NewChunk(),
		err:     nil,
		ttable:  types.NewTypeTable(),
		inFunc:  false,
	}

	c.registerNative("println", -1)

	return c
}

// Adds a constant to the constant pool or, if it's already in it,
// returns the index of the value in the constant pool.
func (c *Compiler) constant(val Const) ByteCode {
	for i, v := range c.constPool {
		if v == val {
			return ByteCode(i)
		}
	}
	c.constPool = append(c.constPool, val)
	return ByteCode(len(c.constPool) - 1)
}

// Get the index of a native function in the natives table.
func (c *Compiler) resolveNative(name string) (ByteCode, bool) {
	for i, n := range c.natives {
		if n.name == name {
			return ByteCode(i), true
		}
	}
	return 0, false
}

// Register a native function in the natives table. -1 arity means
// the function has infinite arguments.
func (c *Compiler) registerNative(name string, arity int) ByteCode {
	if idx, ok := c.resolveNative(name); ok {
		return ByteCode(idx)
	}
	c.natives = append(c.natives, nativeFunc{name: name, arity: arity})
	return ByteCode(len(c.natives) - 1)
}

// Register a new function.
func (c *Compiler) registerFunc(name string, args uint16, locals uint16, chunk *Chunk) ByteCode {
	c.functions = append(c.functions, FuncMeta{
		name:   name,
		args:   args,
		locals: locals,
		chunk:  chunk,
	})
	return ByteCode(len(c.functions) - 1)
}

// Get a function by its name.
func (c *Compiler) resolveFunc(name string) (ByteCode, bool) {
	for i, f := range c.functions {
		if f.name == name {
			return ByteCode(i), true
		}
	}
	return 0, false
}

// Add a new local.
func (c *Compiler) addLocal(name string) ByteCode {
	c.locals = append(c.locals, localVar{name, c.scopeDepth})
	slot := len(c.locals)
	return ByteCode(slot - 1)
}

// Get a local/extern by its name.
func (c *Compiler) resolveVar(name string) (slot ByteCode, ok bool, loadOp OpCode) {
	for i := len(c.locals) - 1; i >= 0; i-- {
		if c.locals[i].name == name && c.locals[i].depth <= c.scopeDepth {
			return ByteCode(i), true, OpLoadLocal
		}
	}

	slot, ok = c.resolveExtern(name)
	loadOp = OpLoadExtern
	return
}

// Add a new extern
func (c *Compiler) addExtern(name string, t *types.Type) ByteCode {
	c.externs = append(c.externs, extern{name, t})
	slot := len(c.externs)
	return ByteCode(slot - 1)
}

// Get an extern symbol by its name.
func (c *Compiler) resolveExtern(name string) (ByteCode, bool) {
	for i := len(c.externs) - 1; i >= 0; i-- {
		if c.externs[i].name == name {
			return ByteCode(i), true
		}
	}

	return 0, false
}

// Begin a new scope.
func (c *Compiler) beginScope() {
	c.scopeDepth++
}

// End the scope.
func (c *Compiler) endScope() {
	c.scopeDepth--
	for len(c.locals) > 0 && c.locals[len(c.locals)-1].depth > c.scopeDepth {
		c.locals = c.locals[:len(c.locals)-1]
	}
}

func (c *Compiler) GetTypes() []*types.TypeDecl {
	return c.ttable.All()
}

func (c *Compiler) Dissasemble() string {
	var str strings.Builder

	str.WriteString("=== const pool ===")
	str.WriteString("\n")
	for i, v := range c.constPool {
		fmt.Fprintf(&str, "[\x1b[35m%d\x1b[0m]: \x1b[33m%v\x1b[0m\n", i, v.Any())
	}

	str.WriteString("\n")
	str.WriteString("=== natives ===")
	str.WriteString("\n")
	for _, native := range c.natives {
		var arityStr string
		if native.arity >= 0 {
			arityStr = fmt.Sprintf("%d", native.arity)
		} else {
			arityStr = "inf"
		}
		fmt.Fprintf(&str, "\x1b[32m%s\x1b[0m(\x1b[33m%s\x1b[0m)\n", native.name, arityStr)
	}

	str.WriteString("\n")
	str.WriteString("=== externs ===")
	str.WriteString("\n")
	for i, extern := range c.externs {
		fmt.Fprintf(&str, "[\x1b[35m%d\x1b[0m]: \x1b[33m%s\x1b[0m \x1b[32m%v\x1b[0m\n", i, extern.t, extern.name)
	}

	str.WriteString("\n")
	str.WriteString("=== functions ===")
	str.WriteString("\n")
	for i, fn := range c.functions {
		fmt.Fprintf(&str, "--- %s at slot %d ---\n", fn.name, i)
		str.WriteString(fn.chunk.String())
		str.WriteString("\n")
	}

	str.WriteString("\n")
	str.WriteString("=== functions (hex) ===")
	str.WriteString("\n")
	for i, fn := range c.functions {
		fmt.Fprintf(&str, "--- %s at slot %d ---\n", fn.name, i)
		str.WriteString(byteDump(fn.chunk.Bytes()))
		str.WriteString("\n")
	}

	return str.String()
}

func (c *Compiler) StringSerialize() string {
	return byteDump(c.Serialize())
}

func byteDump(bytes []byte) string {
	var str strings.Builder

	for i, b := range bytes {
		if b == 0 {
			fmt.Fprintf(&str, "\x1b[90m0\x1b[0m ")
		} else {
			fmt.Fprintf(&str, "\x1b[33m%X\x1b[0m ", b)
		}
		if i > 0 && i%16 == 0 {
			str.WriteString("\n")
		}
	}

	return str.String()
}
