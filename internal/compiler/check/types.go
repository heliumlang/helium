package check

import (
	"fmt"

	"github.com/heliumlang/helium/internal/frontend/parser"
)

type Array struct {
	Value   *Base
	Nesting int
}

type Map struct {
	Key   *Base
	Value *Base
}

type Arg struct {
	Name string
	Type *Base
}

type Closure struct {
	Args    []*Arg
	Returns []*Base
}

type Base struct {
	Name      *string
	TypeArgs  []*TypeInfo
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
	Type    *Base
	Qualifs []string
}

type InterfaceMethod struct {
	Args    []*Arg
	Returns []*Base
}

type Function struct {
	Args     []*Arg
	Returns  []*Base
	Receiver *Base
}

type Initializer struct {
	Args []*Base
}

type Record struct {
	Fields []*Field
}

type Struct struct {
	Fields []*Field
	Inits  []*Initializer
	Impl   []*Base
}

type Interface struct {
	Methods   []*InterfaceMethod
	Constants []*Field
	Generics  []*Base
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
	Type *Base
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

	Base      *Base
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

func (b *Base) Wrap() *TypeInfo {
	return &TypeInfo{
		Kind: TypeBase,
		Base: b,
	}
}

func (b *Base) SetQualifs(optional, throwable bool) {
	b.Optional, b.Throwable = optional, throwable
}

func PlainType(name string) *Base {
	return &Base{
		Name:    &name,
		IsArray: false,
		IsMap:   false,
	}
}

func ArrayType(value *Base, nesting int) *Base {
	return &Base{
		Name:    nil,
		IsArray: true,
		Array: &Array{
			Value:   value,
			Nesting: nesting,
		},
	}
}

func MapType(value, key *Base) *Base {
	return &Base{
		Name:  nil,
		IsMap: true,
		Map: &Map{
			Value: value,
			Key:   key,
		},
	}
}

func NewArg(name string, _type *Base) *Arg {
	return &Arg{
		Name: name,
		Type: _type,
	}
}

func NewFunction(args []*Arg, rets ...*Base) *Function {
	return &Function{
		Args:    args,
		Returns: rets,
	}
}

func ClosureType(args []*Arg, returns ...*Base) *Base {
	return &Base{
		Name:      nil,
		IsClosure: true,
		Closure: &Closure{
			Args:    args,
			Returns: returns,
		},
	}
}

func NewField(name string, _type *Base, qualifs ...string) *Field {
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

func NewInit(args ...*Base) *Initializer {
	return &Initializer{
		Args: args,
	}
}

func FunctionType(args []*Arg, rets ...*Base) *TypeInfo {
	return &TypeInfo{
		Kind:     TypeFunction,
		Function: NewFunction(args, rets...),
	}
}

func StructType(fields []*Field, inits []*Initializer, impl ...*Base) *TypeInfo {
	return &TypeInfo{
		Kind: TypeStruct,
		Struct: &Struct{
			Fields: fields,
			Inits:  inits,
			Impl:   impl,
		},
	}
}

func NewIfaceMethod(args []*Arg, rets ...*Base) *InterfaceMethod {
	return &InterfaceMethod{
		Args:    args,
		Returns: rets,
	}
}

func InterfaceType(functions []*InterfaceMethod, constants []*Field, generics ...*Base) *TypeInfo {
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

func NewVariantCase(name string, _type *Base) *VariantCase {
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

func NodeToType(n parser.Node) *Base {
	switch n := n.(type) {
	case *parser.BaseType:
		plain := PlainType(n.Typename)
		for _, arg := range n.TypeArgs {
			plain.TypeArgs = append(plain.TypeArgs, NodeToType(arg).Wrap())
		}
		plain.SetQualifs(n.Optional, n.Throwable)
		return plain

	case *parser.FunctionType:
		var (
			args []*Arg
			rets []*Base
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
