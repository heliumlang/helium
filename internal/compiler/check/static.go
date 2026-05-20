package check

import (
	"github.com/heliumlang/helium/internal/frontend/parser"
	"github.com/heliumlang/helium/internal/heliumerr"
)

func Check(file string, root parser.Node) (*heliumerr.Error, *TypeTable) {
	table := NewTypeTable()
	return check(file, root, table), table
}

func check(file string, root parser.Node, table *TypeTable) *heliumerr.Error {
	if table == nil {
		table = NewTypeTable()
	}

	var err error

	switch n := root.(type) {
	case *parser.Program:
		for _, item := range n.Items {
			err := check(file, item, table)
			if err != nil {
				return err
			}
		}

	case *parser.Record:
		var fields []*Field
		for _, field := range n.Fields {
			field := field.(*parser.RecordField)
			fields = append(fields, NewField(field.Name, NodeToType(field.Type)))
		}

		record := RecordType(fields...)
		record.SetName(n.Name)

		err = table.Register(record)

	case *parser.Struct:
		var (
			fields []*Field
			inits  []*Initializer
			impls  []*Base
		)
		for _, field := range n.Fields {
			fields = append(fields, NewField(field.Name, NodeToType(field.Type), field.Qualifiers.Slice()...))
		}
		for _, init := range n.Inits {
			var args []*Base
			for _, arg := range init.Params {
				args = append(args, NodeToType(arg.Type))
			}
			inits = append(inits, NewInit(args...))
		}
		for _, impl := range n.Interfaces {
			impls = append(impls, NodeToType(impl))
		}

		_struct := StructType(fields, inits, impls...)
		_struct.SetName(n.Name)

		err = table.Register(_struct)

	case *parser.Interface:
		var (
			methods  []*InterfaceMethod
			consts   []*Field
			generics []*Base
		)
		for _, member := range n.Members {
			switch member := member.(type) {
			case *parser.FnSig:
				var (
					args []*Arg
					rets []*Base
				)
				for _, arg := range member.Args {
					args = append(args, NewArg(arg.Name, NodeToType(arg.Type)))
				}
				for _, ret := range member.Returns {
					rets = append(rets, NodeToType(ret))
				}
				methods = append(methods, NewIfaceMethod(args, rets...))

			case *parser.Const:
				consts = append(consts, NewField(member.Name, NodeToType(member.Type)))
			}
		}
		for _, generic := range n.Generics {
			generics = append(generics, NodeToType(generic))
		}

		iface := InterfaceType(methods, consts, generics...)
		iface.SetName(n.Name)

		err = table.Register(iface)

	case *parser.Enum:
		var cases []*EnumCase
		for _, v := range n.Variants {
			var params []*Arg
			for _, param := range v.Params {
				params = append(params, NewArg(param.Name, NodeToType(param.Type)))
			}
			cases = append(cases, NewEnumCase(v.Name, params...))
		}

		enum := EnumType(cases...)
		enum.SetName(n.Name)

		err = table.Register(enum)

	case *parser.Variant:
		var cases []*VariantCase
		for _, v := range n.Fields {
			cases = append(cases, NewVariantCase(v.Name, NodeToType(v.Type)))
		}

		variant := VariantType(cases...)
		variant.SetName(n.Name)

		err = table.Register(variant)

	case *parser.Alias:
		target := NodeToType(n.Type)
		err = table.Alias(n.Name, target)

	case *parser.FunctionDecl:
		var (
			args []*Arg
			rets []*Base
		)
		for _, arg := range n.Args {
			args = append(args, NewArg(arg.Name, NodeToType(arg.Type)))
		}
		for _, ret := range n.Returns {
			rets = append(rets, NodeToType(ret))
		}

		fn := FunctionType(args, rets...)
		if n.Recv != nil {
			fn.Function.Receiver = NodeToType(n.Recv.Type)
		}
		fn.SetName(n.Name)

		err = table.Register(fn)

	default:
		//return heliumerr.New(fmt.Sprintf("unhandled type: %T", root), heliumerr.EmptyTrace()).SetFilename(file).SetType("semantic")
	}

	if err != nil {
		return heliumerr.New(err.Error(), heliumerr.EmptyTrace()).SetType("semantic").SetFilename(file)
	}
	return nil
}
