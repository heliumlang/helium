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
			t := c.compileExpr(ast.Operand)
			if t == nil {
				return
			}
			if *t.Name == "float" {
				c.current.Emit(OpNegFloat)
			} else {
				c.current.Emit(OpNegInt)
			}

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
			lt := c.compileExpr(ast.Left)
			rt := c.compileExpr(ast.Right)
			if !lt.Equal(rt) {
				c.err = fmt.Errorf("%s and %s can't be %s", *lt.Name, *rt.Name, ast.Op)
				return
			}
			op, _ := c.opForBinary(ast.Op, lt)
			if op == OpNoop {
				c.err = fmt.Errorf("operator %s not defined for type %v", ast.Op, ast.Right)
				return
			}
			c.current.Emit(op)
		}

	case parser.CallExpr:
		c.compileExpr(ast)

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
		c.funcName = ast.Name

		for _, arg := range ast.Args {
			c.addLocal(arg.Name, types.NodeToType(arg.Type))
		}

		for _, n := range ast.Body {
			c.compile(n)
		}
		if len(c.current.code) > 0 {
			if op, _ := c.current.code[len(c.current.code)-1].Decode(); op != OpReturn {
				c.current.Emit(OpReturn, 0)
			}
		} else {
			c.current.Emit(OpReturn, 0)
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
			t := c.compileExpr(ast.Exprs[i])
			if t == nil {
				return
			}
			if ast.Type != nil {
				declared := types.NodeToType(ast.Type)
				if !declared.Equal(t) {
					var name, declName string
					if t.Name != nil {
						name = *t.Name
					} else {
						name = "unknown"
					}
					if declared.Name != nil {
						declName = *declared.Name
					} else {
						declName = "unknown"
					}
					c.err = fmt.Errorf("cannot assign %s to %s", name, declName)
					return
				}
			}
			slot := c.addLocal(name, t)
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
		index := c.current.Last() + 1
		if index == -1 {
			c.err = fmt.Errorf("(INTERNAL) chunk is empty")
			return
		}
		for _, n := range ast.Body {
			c.compile(n)
		}

		t := c.compileExpr(ast.Cond)
		if *t.Name != "bool" {
			c.err = fmt.Errorf("do .. while condition isn't a boolean")
			return
		}
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
			slot := c.addLocal(name, nil)
			c.current.Emit(OpStoreLocal, slot)
		}

		for _, n := range ast.Body {
			c.compile(n)
		}

		c.current.Emit(OpJmp, ByteCode(loopStart))

		c.patchEnd(exitJmp)

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

func (c *Compiler) compileExpr(node parser.Node) *types.Type {
	if c.err != nil {
		return nil
	}

	switch ast := node.(type) {
	case parser.IntLit:
		idx := c.constant(ConstFromInt(ast.Value))
		c.current.Emit(OpPushConst, idx)
		return types.PlainType("int")

	case parser.FloatLit:
		idx := c.constant(ConstFromFloat(ast.Value))
		c.current.Emit(OpPushConst, idx)
		return types.PlainType("float")

	case parser.StringLit:
		idx := c.constant(ConstFromString(ast.Value))
		c.current.Emit(OpPushConst, idx)
		return types.PlainType("string")

	case parser.BoolLit:
		idx := c.constant(ConstFromBool(ast.Value))
		c.current.Emit(OpPushConst, idx)
		return types.PlainType("bool")

	case parser.Ident:
		slot, ok, load := c.resolveVar(ast.Name)
		if !ok {
			c.err = fmt.Errorf("undefined variable: %s", ast.Name)
			return nil
		}
		var t *types.Type
		if load == OpLoadLocal {
			t = c.locals[slot].t
		} else {
			t = c.externs[slot].t
		}
		c.current.Emit(load, slot)
		return t

	case parser.BinaryExpr:
		lt := c.compileExpr(ast.Left)
		rt := c.compileExpr(ast.Right)
		if lt == nil || rt == nil {
			return nil
		}
		if *lt.Name != *rt.Name {
			c.err = fmt.Errorf("type mismatch: %s %s %s", *lt.Name, ast.Op, *rt.Name)
			return nil
		}
		op, ret := c.opForBinary(ast.Op, lt)
		if op == OpNoop {
			c.err = fmt.Errorf("operator %s not defined for %s", ast.Op, *lt.Name)
			return nil
		}
		c.current.Emit(op)
		return ret

	case parser.CallExpr:
		switch callee := ast.Callee.(type) {
		case parser.Ident:
			for _, arg := range ast.Args {
				c.compile(arg.Value)
			}
			if c.inFunc && c.funcName == callee.Name {
				c.current.Emit(OpCall, ByteCode(len(c.functions)), ByteCode(len(ast.Args)))
				return c.ttable.Get(callee.Name).Function.Returns[0]
			} else if idx, ok := c.resolveFunc(callee.Name); ok {
				c.current.Emit(OpCall, idx, ByteCode(len(ast.Args)))
				return c.ttable.Get(callee.Name).Function.Returns[0]
			} else if idx, ok := c.resolveNative(callee.Name); ok {
				c.current.Emit(OpCallNative, idx, ByteCode(len(ast.Args)))
				return c.natives[idx].returns
			} else {
				c.err = fmt.Errorf("undefined function: %s", callee.Name)
			}
		default:
			c.err = fmt.Errorf("unsupported callee: %T", ast.Callee)
		}
	}

	c.err = fmt.Errorf("unhandled expression: %T", node)
	return nil
}

func (c *Compiler) opForBinary(op string, typ *types.Type) (OpCode, *types.Type) {
	if c.err != nil {
		return OpNoop, nil
	}

	var isInt, isFloat bool = false, false
	if typ.Name != nil {
		isInt = *typ.Name == "int"
		isFloat = *typ.Name == "float"
	}
	_bool := types.PlainType("bool")

	switch op {
	case "+":
		if isInt {
			return OpAddInt, typ
		}
		if isFloat {
			return OpAddFloat, typ
		}
	case "-":
		if isInt {
			return OpSubInt, typ
		}
		if isFloat {
			return OpSubFloat, typ
		}
	case "*":
		if isInt {
			return OpMulInt, typ
		}
		if isFloat {
			return OpMulFloat, typ
		}
	case "/":
		if isInt {
			return OpDivInt, typ
		}
		if isFloat {
			return OpDivFloat, typ
		}
	case "==":
		return OpEq, _bool
	case "!=":
		return OpNeq, _bool
	case "<":
		return OpLt, _bool
	case ">":
		return OpGt, _bool
	case "<=":
		return OpLte, _bool
	case ">=":
		return OpGte, _bool
	case "and":
		return OpAnd, _bool
	case "or":
		return OpOr, _bool
	}
	return OpNoop, nil
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
		rt := c.compileExpr(ast.Right)

		switch ast.Op {
		case "+=":
			op, _ := c.opForBinary("+", rt)
			c.current.Emit(op)
		case "-=":
			op, _ := c.opForBinary("-", rt)
			c.current.Emit(op)
		case "*=":
			op, _ := c.opForBinary("*", rt)
			c.current.Emit(op)
		case "/=":
			op, _ := c.opForBinary("/", rt)
			c.current.Emit(op)
		case "%=":
			op, _ := c.opForBinary("%", rt)
			c.current.Emit(op)

		}
	} else {
		c.compile(ast.Right)
	}

	c.current.Emit(OpStoreLocal, slot)
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
