package compiler

import (
	"fmt"
	"math"
)

type ConstType int

const (
	ConstInt ConstType = iota
	ConstFloat
	ConstString
	ConstChar
	ConstBool
)

type Const struct {
	Type ConstType
	i    int64
	f    float64
	str  string
	ch   byte
	b    bool
}

func ConstFromInt(v int64) Const {
	return Const{Type: ConstInt, i: v}
}

func ConstFromFloat(v float64) Const {
	return Const{Type: ConstFloat, f: v}
}

func ConstFromString(v string) Const {
	return Const{Type: ConstString, str: v}
}

func ConstFromChar(v byte) Const {
	return Const{Type: ConstChar, ch: v}
}

func ConstFromBool(v bool) Const {
	return Const{Type: ConstBool, b: v}
}

func (c Const) Int() int64     { return c.i }
func (c Const) Float() float64 { return c.f }
func (c Const) Str() string    { return c.str }
func (c Const) Char() byte     { return c.ch }
func (c Const) Bool() bool     { return c.b }

// Serialize a single constant to bytes.
func (c Const) Bytes() []byte {
	var buf []byte
	buf = append(buf, byte(c.Type))
	switch c.Type {
	case ConstInt:
		buf = appendI64(buf, c.i)
	case ConstFloat:
		bits := math.Float64bits(c.f)
		buf = appendU64(buf, bits)
	case ConstString:
		buf = appendU16(buf, uint16(len(c.str)))
		buf = append(buf, []byte(c.str)...)
	case ConstChar:
		buf = append(buf, c.ch)
	case ConstBool:
		if c.b {
			buf = append(buf, 1)
		} else {
			buf = append(buf, 0)
		}
	}
	return buf
}

// Get the constant as any.
func (c Const) Value() any {
	switch c.Type {
	case ConstInt:
		return c.i
	case ConstFloat:
		return c.f
	case ConstString:
		return c.str
	case ConstChar:
		return c.ch
	case ConstBool:
		return c.b
	}

	return nil
}

// Get the constant as a string.
func (c Const) String() string {
	switch c.Type {
	case ConstInt:
		return fmt.Sprintf("\x1b[33m%d\x1b[0m", c.i)
	case ConstFloat:
		return fmt.Sprintf("\x1b[33m%f\x1b[0m", c.f)
	case ConstString:
		return fmt.Sprintf("\x1b[32m%q\x1b[0m", c.str)
	case ConstChar:
		return fmt.Sprintf("\x1b[35m'%c'\x1b[0m", c.ch)
	case ConstBool:
		return fmt.Sprintf("\x1b[37m%t\x1b[0m", c.b)
	}
	return ""
}
