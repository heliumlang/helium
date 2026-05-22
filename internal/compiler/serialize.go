package compiler

func (c *Compiler) Serialize() []byte {
	var buf []byte

	// module
	buf = appendU16(buf, uint16(len(c.module)))
	buf = append(buf, []byte(c.module)...)

	// const pool
	buf = appendU16(buf, uint16(len(c.constPool)))
	for _, con := range c.constPool {
		buf = append(buf, con.Bytes()...)
	}

	// native table
	buf = appendU16(buf, uint16(len(c.natives)))
	for _, name := range c.natives {
		buf = appendU16(buf, uint16(len(name)))
		buf = append(buf, []byte(name)...)
	}

	// function table
	buf = appendU16(buf, uint16(len(c.functions)))
	for _, fn := range c.functions {
		buf = appendU16(buf, uint16(len(fn.name)))
		buf = append(buf, []byte(fn.name)...)
		buf = appendU16(buf, fn.args)
		buf = appendU16(buf, fn.locals)
		code := fn.chunk.Bytes()
		buf = appendU16(buf, uint16(len(code)))
		buf = append(buf, code...)
	}

	return buf
}

func appendU16(buf []byte, v uint16) []byte {
	return append(buf, byte(v>>8), byte(v))
}

func appendU64(buf []byte, v uint64) []byte {
	return append(buf,
		byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
		byte(v>>24), byte(v>>16), byte(v>>8), byte(v),
	)
}

func appendI64(buf []byte, v int64) []byte {
	return appendU64(buf, uint64(v))
}
