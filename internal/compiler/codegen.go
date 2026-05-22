package compiler

import (
	"fmt"

	"github.com/heliumlang/helium/internal/frontend/parser"
)

const maxU16 = ^uint16(0)

func (c *Compiler) Compile(ast parser.Node) error {
	c.compile(ast)
	return c.err
}

func (c *Compiler) compile(ast parser.Node) {
	if c.err != nil {
		return
	}

	switch ast := ast.(type) {
	case parser.IntLit:
		v := c.constant(ConstFromInt(ast.Value))
		c.current.Emit(OpPushConst, v)

	case parser.FloatLit:
		v := c.constant(ConstFromFloat(ast.Value))
		c.current.Emit(OpPushConst, v)

	case parser.StringLit:
		v := c.constant(ConstFromString(ast.Value))
		c.current.Emit(OpPushConst, v)

	case parser.CharLit:
		v := c.constant(ConstFromChar(ast.Value))
		c.current.Emit(OpPushConst, v)

	case parser.BoolLit:
		v := c.constant(ConstFromBool(ast.Value))
		c.current.Emit(OpPushConst, v)

	case parser.NoneLit:
		c.current.Emit(OpPushNone)

	case parser.ArrayLit:
		count := len(ast.Elements)
		if count > int(maxU16) {
			c.err = fmt.Errorf("array elements exceed the %d limit", maxU16)
		}

		for _, elem := range ast.Elements {
			c.compile(elem)
		}

		c.current.Emit(OpPushArr, ByteCode(count))

	case parser.MapLit:
		count := len(ast.Pairs)
		if count > int(maxU16) {
			c.err = fmt.Errorf("map pairs exceed the %d limit", maxU16)
		}

		for _, pair := range ast.Pairs {
			c.compile(pair.Key)
			c.Compile(pair.Value)
		}

		c.current.Emit(OpPushMap, ByteCode(count))

	case parser.Ident:
		local, ok := c.resolveLocal(ast.Name)
		if !ok {
			c.err = fmt.Errorf("unidentified variable: %s", ast.Name)
		}
		c.current.Emit(OpLoadLocal, local)

	case parser.UnaryExpr:
		c.compile(ast.Operand)

		switch ast.Op {
		case "!":
			c.current.Emit(OpNot)

		case "-":
			c.current.Emit(OpNeg)

		case "++":
			c.current.Emit(OpInc)

		case "--":
			c.current.Emit(OpDec)
		}

	case parser.BinaryExpr:
		switch ast.Op {
		case "=", "+=", "-=", "*=", "/=", "%=":
			c.compileAssign(ast)

		default:
			c.compile(ast.Left)
			c.compile(ast.Right)
			c.current.Emit(binaryExprTable[ast.Op])
		}

	default:
		c.err = fmt.Errorf("unhandled type")
	}
}

var binaryExprTable = map[string]OpCode{
	"+": OpAdd, "-": OpSub, "*": OpMul, "/": OpDiv, "%": OpMod,
	"or": OpOr, "and": OpAnd,
	"==": OpEq, "!=": OpNeq, ">": OpGt, "<": OpLt, ">=": OpGte, "<=": OpLte,
}

func (c *Compiler) compileAssign(ast parser.BinaryExpr) error {
	slot, ok := c.resolveLocal(ast.Left.(parser.Ident).Name)
	if !ok {
		return fmt.Errorf("undefined variable: %s", ast.Left.(parser.Ident).Name)
	}

	if ast.Op != "=" {
		c.current.Emit(OpLoadLocal, slot)
		c.compile(ast.Right)

		switch ast.Op {
		case "+=":
			c.current.Emit(OpAdd)
		case "-=":
			c.current.Emit(OpSub)
		case "*=":
			c.current.Emit(OpMul)
		case "/=":
			c.current.Emit(OpDiv)
		case "%=":
			c.current.Emit(OpMod)
		}
	} else {
		c.compile(ast.Right)
	}

	c.current.Emit(OpStoreLocal, slot)
	return nil
}
