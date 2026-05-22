package compiler

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

type Compiler struct {
	module     string
	current    *Chunk // current chunk being written to
	constPool  []Const
	functions  []FuncMeta
	natives    []string // native function names
	locals     []localVar
	scopeDepth int
	err        error
}

func NewCompiler(module string) *Compiler {
	return &Compiler{
		module:  module,
		current: NewChunk(),
		err:     nil,
	}
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
		if n == name {
			return ByteCode(i), true
		}
	}
	return 0, false
}

// Register a native function in the natives table.
func (c *Compiler) registerNative(name string) ByteCode {
	if idx, ok := c.resolveNative(name); ok {
		return ByteCode(idx)
	}
	c.natives = append(c.natives, name)
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
	return ByteCode(len(c.locals))
}

// Get a local by its name.
func (c *Compiler) resolveLocal(name string) (ByteCode, bool) {
	for i := len(c.locals) - 1; i >= 0; i-- {
		if c.locals[i].name == name {
			return ByteCode(i), true
		}
	}
	return 0, false
}

// Begin a new scope.
func (c *Compiler) beginScope() {
	c.scopeDepth++
}

// End the scope, thus removing all locals of the scope.
func (c *Compiler) endScope() {
	c.scopeDepth--
	for len(c.locals) > 0 && c.locals[len(c.locals)-1].depth > c.scopeDepth {
		c.locals = c.locals[:len(c.locals)-1]
	}
}
