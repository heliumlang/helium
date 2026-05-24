package compiler

import (
	"fmt"

	"github.com/heliumlang/helium/internal/frontend/parser"
	"github.com/heliumlang/helium/internal/types"
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
	case *parser.Module:
		c.module = ast.Name

	case parser.Extern:
		for _, member := range ast.Members {
			c.addExtern(member.Name, types.NodeToType(member.Type))
		}

	case parser.Noop:
		c.current.Emit(OpNoop)

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
		local, ok, load := c.resolveVar(ast.Name)
		if !ok {
			c.err = fmt.Errorf("unidentified variable: %s", ast.Name)
		}
		c.current.Emit(load, local)

	case parser.UnaryExpr:

		switch ast.Op {
		case "!":
			c.compile(ast.Operand)
			c.current.Emit(OpNot)

		case "-":
			c.compile(ast.Operand)
			c.current.Emit(OpNeg)

		case "++":
			name := ast.Operand.(parser.Ident).Name
			slot, ok, _ := c.resolveVar(name)
			if !ok {
				c.err = fmt.Errorf("undefined variable: %s", name)
			}
			c.current.Emit(OpInc, slot)

		case "--":
			name := ast.Operand.(parser.Ident).Name
			slot, ok, _ := c.resolveVar(name)
			if !ok {
				c.err = fmt.Errorf("undefined variable: %s", name)
			}
			c.current.Emit(OpDec, slot)
		}

	case parser.BinaryExpr:
		switch ast.Op {
		case "=", "+=", "-=", "*=", "/=", "%=":
			if err := c.compileAssign(ast); err != nil {
				c.err = err
				return
			}

		default:
			c.compile(ast.Left)
			c.compile(ast.Right)
			c.current.Emit(binaryExprTable[ast.Op])
		}

	case *parser.Program:
		for _, item := range ast.Items {
			c.compile(item)
		}

	case *parser.Record:
		var fields []*types.Field
		for _, field := range ast.Fields {
			field := field.(*parser.RecordField)
			fields = append(fields, types.NewField(field.Name, types.NodeToType(field.Type)))
		}

		record := types.RecordType(fields...)
		record.SetName(ast.Name)

		c.err = c.ttable.Register(record)

	case *parser.Struct:
		var (
			fields []*types.Field
			inits  []*types.Initializer
			impls  []*types.Type
		)
		for _, field := range ast.Fields {
			fields = append(fields, types.NewField(field.Name, types.NodeToType(field.Type), field.Qualifiers.Slice()...))
		}
		for _, init := range ast.Inits {
			var args []*types.Type
			for _, arg := range init.Params {
				args = append(args, types.NodeToType(arg.Type))
			}
			inits = append(inits, types.NewInit(args...))
		}
		for _, impl := range ast.Interfaces {
			impls = append(impls, types.NodeToType(impl))
		}

		_struct := types.StructType(fields, inits, impls...)
		_struct.SetName(ast.Name)

		c.err = c.ttable.Register(_struct)

	case *parser.Interface:
		var (
			methods  []*types.InterfaceMethod
			consts   []*types.Field
			generics []*types.Type
		)
		for _, member := range ast.Members {
			switch member := member.(type) {
			case *parser.FnSig:
				var (
					args []*types.Arg
					rets []*types.Type
				)
				for _, arg := range member.Args {
					args = append(args, types.NewArg(arg.Name, types.NodeToType(arg.Type)))
				}
				for _, ret := range member.Returns {
					rets = append(rets, types.NodeToType(ret))
				}
				methods = append(methods, types.NewIfaceMethod(args, rets...))

			case *parser.Const:
				consts = append(consts, types.NewField(member.Name, types.NodeToType(member.Type)))
			}
		}
		for _, generic := range ast.Generics {
			generics = append(generics, types.NodeToType(generic))
		}

		iface := types.InterfaceType(methods, consts, generics...)
		iface.SetName(ast.Name)

		c.err = c.ttable.Register(iface)

	case *parser.Enum:
		var cases []*types.EnumCase
		for _, v := range ast.Variants {
			var params []*types.Arg
			for _, param := range v.Params {
				params = append(params, types.NewArg(param.Name, types.NodeToType(param.Type)))
			}
			cases = append(cases, types.NewEnumCase(v.Name, params...))
		}

		enum := types.EnumType(cases...)
		enum.SetName(ast.Name)

		c.err = c.ttable.Register(enum)

	case *parser.Variant:
		var cases []*types.VariantCase
		for _, v := range ast.Fields {
			cases = append(cases, types.NewVariantCase(v.Name, types.NodeToType(v.Type)))
		}

		variant := types.VariantType(cases...)
		variant.SetName(ast.Name)

		c.err = c.ttable.Register(variant)

	case *parser.Alias:
		target := types.NodeToType(ast.Type)
		c.err = c.ttable.Alias(ast.Name, target)

	case *parser.FunctionDecl:
		var (
			args []*types.Arg
			rets []*types.Type
		)
		for _, arg := range ast.Args {
			args = append(args, types.NewArg(arg.Name, types.NodeToType(arg.Type)))
		}
		for _, ret := range ast.Returns {
			rets = append(rets, types.NodeToType(ret))
		}

		fn := types.FunctionType(args, rets...)
		if ast.Recv != nil {
			fn.Function.Receiver = types.NodeToType(ast.Recv.Type)
		}
		fn.SetName(ast.Name)

		c.err = c.ttable.Register(fn)

		prevChunk := c.current
		prevLocals := c.locals
		prevDepth := c.scopeDepth

		c.beginScope()
		c.inFunc = true

		chunk := NewChunk()
		c.current = chunk
		c.locals = nil
		c.scopeDepth = 0

		for _, arg := range ast.Args {
			c.addLocal(arg.Name)
		}

		for _, n := range ast.Body {
			c.compile(n)
		}

		locals := len(c.locals) - len(ast.Args)

		c.current = prevChunk
		c.locals = prevLocals
		c.scopeDepth = prevDepth

		c.endScope()
		c.inFunc = false
		c.registerFunc(ast.Name, uint16(len(ast.Args)), uint16(locals), chunk)

	case parser.Return:
		for _, expr := range ast.Exprs {
			c.compile(expr)
		}
		count := len(ast.Exprs)
		if count > int(maxU16) {
			c.err = fmt.Errorf("can't return more than %d values", maxU16)
			return
		}
		c.current.Emit(OpReturn, ByteCode(count))

	case *parser.VarDecl:
		if !c.inFunc {
			c.err = fmt.Errorf("locals must go inside functions")
			return
		}

		if len(ast.Exprs) > len(ast.Idents) {
			c.err = fmt.Errorf("more expressions than names in variable declaration")
			return
		} else if len(ast.Idents) > len(ast.Exprs) {
			c.err = fmt.Errorf("more names than expressions in variable declaration")
			return
		}

		for i, name := range ast.Idents {
			slot := c.addLocal(name)
			c.compile(ast.Exprs[i])
			c.current.Emit(OpStoreLocal, slot)
		}

	case parser.ExprStmt:
		c.compile(ast.Expr)

	case parser.IfStmt:
		var endJmps []int

		c.compile(ast.Cond)
		ifJmp := c.current.Emit(OpJmpIfFalse, 0)
		for _, n := range ast.Body {
			c.compile(n)
		}
		endJmps = append(endJmps, c.current.Emit(OpJmp, 0))
		c.patchEnd(ifJmp)

		for _, elif := range ast.Elifs {
			c.compile(elif.Cond)
			elifJmp := c.current.Emit(OpJmpIfFalse, 0)
			for _, n := range elif.Body {
				c.compile(n)
			}
			endJmps = append(endJmps, c.current.Emit(OpJmp, 0))
			c.patchEnd(elifJmp)
		}

		if ast.Else != nil {
			for _, n := range ast.Else {
				c.compile(n)
			}
		}

		for _, jmp := range endJmps {
			c.patchEnd(jmp)
		}

	case parser.Do:
		index := c.current.Last()
		if index == -1 {
			c.err = fmt.Errorf("(INTERNAL) chunk is empty")
			return
		}
		for _, n := range ast.Body {
			c.compile(n)
		}

		c.compile(ast.Cond)
		c.current.Emit(OpJmpIfTrue, ByteCode(index))

	case parser.For:
		c.compile(ast.Iter)
		c.current.Emit(OpMakeIter)

		loopStart := c.current.Last()
		if loopStart == -1 {
			c.err = fmt.Errorf("(INTERNAL) chunk is empty")
			return
		}
		exitJmp := c.current.Emit(OpNextIter, 0)

		for _, name := range ast.Idents {
			slot := c.addLocal(name)
			c.current.Emit(OpStoreLocal, slot)
		}

		for _, n := range ast.Body {
			c.compile(n)
		}

		c.current.Emit(OpJmp, ByteCode(loopStart))

		c.patchEnd(exitJmp)

	case parser.CallExpr:
		switch callee := ast.Callee.(type) {
		case parser.Ident:
			for _, arg := range ast.Args {
				c.compile(arg.Value)
			}
			if idx, ok := c.resolveFunc(callee.Name); ok {
				c.current.Emit(OpCall, idx, ByteCode(len(ast.Args)))
			} else if idx, ok := c.resolveNative(callee.Name); ok {
				c.current.Emit(OpCallNative, idx, ByteCode(len(ast.Args)))
			} else {
				c.err = fmt.Errorf("undefined function: %s", callee.Name)
			}
		default:
			c.err = fmt.Errorf("unsupported callee: %T", ast.Callee)
		}

		// case parser.MethodCall:
		//     c.compile(ast.Object)
		//     for _, arg := range ast.Args {
		//         c.compile(arg)
		//     }
		//     c.current.Emit(OpCallMethod, ...)

		// case parser.IndexExpr:
		//     c.compile(ast.Object)
		//     c.compile(ast.Index)
		//     c.current.Emit(OpIndex)

		// case parser.FieldAccess:
		//     c.compile(ast.Object)
		//     // field name needs to be in the const pool as a string
		//     idx := c.constant(ConstFromString(ast.Field))
		//     c.current.Emit(OpGetField, idx)

		// case parser.ForceUnwrap:
		//     c.compile(ast.Operand)
		//     c.current.Emit(OpForceUnwrap)

		// case parser.OptionalChain:
		//     c.compile(ast.Operand)
		//     c.current.Emit(OpOptionalChain)

	default:
		c.err = fmt.Errorf("unhandled type %T", ast)
	}
}

var binaryExprTable = map[string]OpCode{
	"+": OpAdd, "-": OpSub, "*": OpMul, "/": OpDiv, "%": OpMod,
	"or": OpOr, "and": OpAnd,
	"==": OpEq, "!=": OpNeq, ">": OpGt, "<": OpLt, ">=": OpGte, "<=": OpLte,
}

func (c *Compiler) compileAssign(ast parser.BinaryExpr) error {
	slot, ok, load := c.resolveVar(ast.Left.(parser.Ident).Name)
	if load == OpLoadExtern {
		return fmt.Errorf("can't assign a value to an extern member")
	}
	if !ok {
		return fmt.Errorf("undefined variable: %s", ast.Left.(parser.Ident).Name)
	}

	if ast.Op != "=" {
		c.current.Emit(load, slot)
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

	c.current.Emit(OpLoadLocal, slot)
	return nil
}

func (c *Compiler) patchEnd(index int) {
	if err := c.current.PatchToEnd(index); err != nil {
		c.err = err
	}
}

func (c *Compiler) patch(index int, target int) {
	if err := c.current.Patch(index, target); err != nil {
		c.err = err
	}
}
