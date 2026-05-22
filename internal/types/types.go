package types

import (
	"fmt"

	"github.com/heliumlang/helium/internal/frontend/parser"
)

type Array struct {
	Value   *BaseType
	Nesting int
}

type Map struct {
	Key   *BaseType
	Value *BaseType
}

type Arg struct {
	Name string
	Type *BaseType
}

type Closure struct {
	Args    []*Arg
	Returns []*BaseType
}

type BaseType struct {
	Name      *string
	TypeArgs  []*BaseType
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
	Type    *BaseType
	Qualifs []string
}

type InterfaceMethod struct {
	Args    []*Arg
	Returns []*BaseType
}

type Function struct {
	Args     []*Arg
	Returns  []*BaseType
	Receiver *BaseType
}

type Initializer struct {
	Args []*BaseType
}

type Record struct {
	Fields []*Field
}

type Struct struct {
	Fields []*Field
	Inits  []*Initializer
	Impl   []*BaseType
}

type Interface struct {
	Methods   []*InterfaceMethod
	Constants []*Field
	Generics  []*BaseType
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
	Type *BaseType
}

type Variant struct {
	Cases []*VariantCase
}

type TypeKind int

const (
	TypeBase TypeKind = iota
	TypeRecord
	TypeStruct
	TypeInterface
	TypeEnum
	TypeVariant
	TypeFunction
)

type TypeInfo struct {
	name    string
	setName bool
	Kind    TypeKind

	Base      *BaseType
	Record    *Record
	Struct    *Struct
	Interface *Interface
	Enum      *Enum
	Variant   *Variant
	Function  *Function
}

func (ti *TypeInfo) SetName(name string) {
	ti.setName = true
	ti.name = name
}

func (ti *TypeInfo) HasName() bool {
	return ti.setName
}

func (ti *TypeInfo) GetName() string {
	return ti.name
}

func RawType(kind TypeKind) *TypeInfo {
	return &TypeInfo{Kind: kind}
}

func (b *BaseType) Wrap() *TypeInfo {
	return &TypeInfo{
		Kind: TypeBase,
		Base: b,
	}
}

func (b *BaseType) SetOptional(v bool) {
	b.Optional = v
}

func (b *BaseType) SetThrowable(v bool) {
	b.Throwable = v
}

func (b *BaseType) SetQualifs(optional, throwable bool) {
	b.SetOptional(optional)
	b.SetThrowable(throwable)
}

func PlainType(name string) *BaseType {
	return &BaseType{
		Name:    &name,
		IsArray: false,
		IsMap:   false,
	}
}

func ArrayType(value *BaseType, nesting int) *BaseType {
	return &BaseType{
		Name:    nil,
		IsArray: true,
		Array: &Array{
			Value:   value,
			Nesting: nesting,
		},
	}
}

func MapType(value, key *BaseType) *BaseType {
	return &BaseType{
		Name:  nil,
		IsMap: true,
		Map: &Map{
			Value: value,
			Key:   key,
		},
	}
}

func NewArg(name string, _type *BaseType) *Arg {
	return &Arg{
		Name: name,
		Type: _type,
	}
}

func NewFunction(args []*Arg, rets ...*BaseType) *Function {
	return &Function{
		Args:    args,
		Returns: rets,
	}
}

func ClosureType(args []*Arg, returns ...*BaseType) *BaseType {
	return &BaseType{
		Name:      nil,
		IsClosure: true,
		Closure: &Closure{
			Args:    args,
			Returns: returns,
		},
	}
}

func NewField(name string, _type *BaseType, qualifs ...string) *Field {
	return &Field{
		Name:    name,
		Type:    _type,
		Qualifs: qualifs,
	}
}

func RecordType(fields ...*Field) *TypeInfo {
	return &TypeInfo{
		Kind: TypeRecord,
		Record: &Record{
			Fields: fields,
		},
	}
}

func NewInit(args ...*BaseType) *Initializer {
	return &Initializer{
		Args: args,
	}
}

func FunctionType(args []*Arg, rets ...*BaseType) *TypeInfo {
	return &TypeInfo{
		Kind:     TypeFunction,
		Function: NewFunction(args, rets...),
	}
}

func StructType(fields []*Field, inits []*Initializer, impl ...*BaseType) *TypeInfo {
	return &TypeInfo{
		Kind: TypeStruct,
		Struct: &Struct{
			Fields: fields,
			Inits:  inits,
			Impl:   impl,
		},
	}
}

func NewIfaceMethod(args []*Arg, rets ...*BaseType) *InterfaceMethod {
	return &InterfaceMethod{
		Args:    args,
		Returns: rets,
	}
}

func InterfaceType(functions []*InterfaceMethod, constants []*Field, generics ...*BaseType) *TypeInfo {
	return &TypeInfo{
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

func EnumType(cases ...*EnumCase) *TypeInfo {
	return &TypeInfo{
		Kind: TypeEnum,
		Enum: &Enum{
			Cases: cases,
		},
	}
}

func NewVariantCase(name string, _type *BaseType) *VariantCase {
	return &VariantCase{
		Name: name,
		Type: _type,
	}
}

func VariantType(cases ...*VariantCase) *TypeInfo {
	return &TypeInfo{
		Kind: TypeVariant,
		Variant: &Variant{
			Cases: cases,
		},
	}
}

func NodeToType(n parser.Node) *BaseType {
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
			rets []*BaseType
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
