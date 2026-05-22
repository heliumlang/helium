package types

import (
	"fmt"

	"github.com/heliumlang/helium/internal/frontend/parser"
)

type Array struct {
	Value   *Type
	Nesting int
}

type Map struct {
	Key   *Type
	Value *Type
}

type Arg struct {
	Name string
	Type *Type
}

type Closure struct {
	Args    []*Arg
	Returns []*Type
}

type Type struct {
	Name      *string
	TypeArgs  []*Type
	Optional  bool
	Throwable bool

	IsArray bool
	Array   *Array

	IsMap bool
	Map   *Map

	IsClosure bool
	Closure   *Closure
}

type Field struct {
	Name    string
	Type    *Type
	Qualifs []string
}

type InterfaceMethod struct {
	Args    []*Arg
	Returns []*Type
}

type Function struct {
	Args     []*Arg
	Returns  []*Type
	Receiver *Type
}

type Initializer struct {
	Args []*Type
}

type Record struct {
	Fields []*Field
}

type Struct struct {
	Fields []*Field
	Inits  []*Initializer
	Impl   []*Type
}

type Interface struct {
	Methods   []*InterfaceMethod
	Constants []*Field
	Generics  []*Type
}

type EnumCase struct {
	Name string
	Args []*Arg
}

type Enum struct {
	Cases []*EnumCase
}

type VariantCase struct {
	Name string
	Type *Type
}

type Variant struct {
	Cases []*VariantCase
}

type TypeKind int

const (
	TypeRecord TypeKind = iota
	TypeStruct
	TypeInterface
	TypeEnum
	TypeVariant
	TypeFunction
)

type TypeDecl struct {
	name    string
	setName bool
	Kind    TypeKind

	Record    *Record
	Struct    *Struct
	Interface *Interface
	Enum      *Enum
	Variant   *Variant
	Function  *Function
}

func (ti *TypeDecl) SetName(name string) {
	ti.setName = true
	ti.name = name
}

func (ti *TypeDecl) HasName() bool {
	return ti.setName
}

func (ti *TypeDecl) GetName() string {
	return ti.name
}

func RawType(kind TypeKind) *TypeDecl {
	return &TypeDecl{Kind: kind}
}

func (b *Type) SetOptional(v bool) {
	b.Optional = v
}

func (b *Type) SetThrowable(v bool) {
	b.Throwable = v
}

func (b *Type) SetQualifs(optional, throwable bool) {
	b.SetOptional(optional)
	b.SetThrowable(throwable)
}

func PlainType(name string) *Type {
	return &Type{
		Name:    &name,
		IsArray: false,
		IsMap:   false,
	}
}

func ArrayType(value *Type, nesting int) *Type {
	return &Type{
		Name:    nil,
		IsArray: true,
		Array: &Array{
			Value:   value,
			Nesting: nesting,
		},
	}
}

func MapType(value, key *Type) *Type {
	return &Type{
		Name:  nil,
		IsMap: true,
		Map: &Map{
			Value: value,
			Key:   key,
		},
	}
}

func NewArg(name string, _type *Type) *Arg {
	return &Arg{
		Name: name,
		Type: _type,
	}
}

func NewFunction(args []*Arg, rets ...*Type) *Function {
	return &Function{
		Args:    args,
		Returns: rets,
	}
}

func ClosureType(args []*Arg, returns ...*Type) *Type {
	return &Type{
		Name:      nil,
		IsClosure: true,
		Closure: &Closure{
			Args:    args,
			Returns: returns,
		},
	}
}

func NewField(name string, _type *Type, qualifs ...string) *Field {
	return &Field{
		Name:    name,
		Type:    _type,
		Qualifs: qualifs,
	}
}

func RecordType(fields ...*Field) *TypeDecl {
	return &TypeDecl{
		Kind: TypeRecord,
		Record: &Record{
			Fields: fields,
		},
	}
}

func NewInit(args ...*Type) *Initializer {
	return &Initializer{
		Args: args,
	}
}

func FunctionType(args []*Arg, rets ...*Type) *TypeDecl {
	return &TypeDecl{
		Kind:     TypeFunction,
		Function: NewFunction(args, rets...),
	}
}

func StructType(fields []*Field, inits []*Initializer, impl ...*Type) *TypeDecl {
	return &TypeDecl{
		Kind: TypeStruct,
		Struct: &Struct{
			Fields: fields,
			Inits:  inits,
			Impl:   impl,
		},
	}
}

func NewIfaceMethod(args []*Arg, rets ...*Type) *InterfaceMethod {
	return &InterfaceMethod{
		Args:    args,
		Returns: rets,
	}
}

func InterfaceType(functions []*InterfaceMethod, constants []*Field, generics ...*Type) *TypeDecl {
	return &TypeDecl{
		Kind: TypeInterface,
		Interface: &Interface{
			Methods:   functions,
			Constants: constants,
			Generics:  generics,
		},
	}
}

func NewEnumCase(name string, args ...*Arg) *EnumCase {
	return &EnumCase{
		Name: name,
		Args: args,
	}
}

func EnumType(cases ...*EnumCase) *TypeDecl {
	return &TypeDecl{
		Kind: TypeEnum,
		Enum: &Enum{
			Cases: cases,
		},
	}
}

func NewVariantCase(name string, _type *Type) *VariantCase {
	return &VariantCase{
		Name: name,
		Type: _type,
	}
}

func VariantType(cases ...*VariantCase) *TypeDecl {
	return &TypeDecl{
		Kind: TypeVariant,
		Variant: &Variant{
			Cases: cases,
		},
	}
}

func NodeToType(n parser.Node) *Type {
	switch n := n.(type) {
	case *parser.BaseType:
		plain := PlainType(n.Typename)
		for _, arg := range n.TypeArgs {
			plain.TypeArgs = append(plain.TypeArgs, NodeToType(arg))
		}
		plain.SetQualifs(n.Optional, n.Throwable)
		return plain

	case *parser.FunctionType:
		var (
			args []*Arg
			rets []*Type
		)

		for _, arg := range n.Args {
			arg := arg.(*parser.DeclArg)
			args = append(args, NewArg(arg.Name, NodeToType(arg.Type)))
		}
		for _, ret := range n.Returns {
			rets = append(rets, NodeToType(ret))
		}
		return ClosureType(args, rets...)

	case *parser.ArrayType:
		nesting := arrayNesting(n)
		arr := ArrayType(NodeToType(n.Values), nesting)
		arr.SetQualifs(n.Optional, n.Throwable)
		return arr

	case *parser.MapType:
		m := MapType(NodeToType(n.Value), NodeToType(n.Key))
		m.SetQualifs(n.Optional, n.Throwable)
		return m

	default:
		fmt.Println("returned nil in NodeToType")
		return nil
	}
}

func arrayNesting(n *parser.ArrayType) int {
	nesting := 0
	var current parser.Node = n
	for {
		arr, ok := current.(*parser.ArrayType)
		if !ok {
			break
		}
		nesting++
		current = arr.Values
	}
	return nesting
}
