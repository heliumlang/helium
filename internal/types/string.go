package types

import (
	"fmt"
	"strings"
)

func (b *Type) String() string {
	if b == nil {
		return "<nil>"
	}

	var sb strings.Builder

	switch {
	case b.IsArray:
		sb.WriteString(b.Array.String())

	case b.IsMap:
		sb.WriteString(b.Map.String())

	case b.IsClosure:
		sb.WriteString(b.Closure.String())

	case b.Name != nil:
		sb.WriteString(*b.Name)
		if len(b.TypeArgs) > 0 {
			sb.WriteString("<")
			for i, arg := range b.TypeArgs {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(arg.String())
			}
			sb.WriteString(">")
		}

	default:
		sb.WriteString("<unknown>")
	}

	if b.Optional {
		sb.WriteString("?")
	}
	if b.Throwable {
		sb.WriteString("!")
	}

	return sb.String()
}

func (a *Array) String() string {
	if a == nil {
		return "<nil>"
	}
	return a.Value.String() + strings.Repeat("[]", a.Nesting)
}

func (m *Map) String() string {
	if m == nil {
		return "<nil>"
	}
	return fmt.Sprintf("[%s]%s", m.Key.String(), m.Value.String())
}

func (a *Arg) String() string {
	if a == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s: %s", a.Name, a.Type.String())
}

func (c *Closure) String() string {
	if c == nil {
		return "<nil>"
	}

	var sb strings.Builder
	sb.WriteString("(")

	for i, arg := range c.Args {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(arg.String())
	}

	sb.WriteString(")")

	switch len(c.Returns) {
	case 0:
	case 1:
		sb.WriteString(" -> ")
		sb.WriteString(c.Returns[0].String())
	default:
		sb.WriteString(" -> (")
		for i, r := range c.Returns {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(r.String())
		}
		sb.WriteString(")")
	}

	return sb.String()
}

func (f *Field) String() string {
	if f == nil {
		return "<nil>"
	}

	var sb strings.Builder

	if len(f.Qualifs) > 0 {
		sb.WriteString(strings.Join(f.Qualifs, " "))
		sb.WriteString(" ")
	}

	sb.WriteString(f.Name)
	sb.WriteString(": ")
	sb.WriteString(f.Type.String())

	return sb.String()
}

func (fn *Function) String() string {
	if fn == nil {
		return "<nil>"
	}

	var sb strings.Builder

	if fn.Receiver != nil {
		sb.WriteString("[")
		sb.WriteString(fn.Receiver.String())
		sb.WriteString("]")
	}

	sb.WriteString("(")

	for i, arg := range fn.Args {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(arg.String())
	}

	sb.WriteString(")")

	switch len(fn.Returns) {
	case 0:
	case 1:
		sb.WriteString(" -> ")
		sb.WriteString(fn.Returns[0].String())
	default:
		sb.WriteString(" -> (")
		for i, r := range fn.Returns {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(r.String())
		}
		sb.WriteString(")")
	}

	return sb.String()
}

func (fn *InterfaceMethod) String() string {
	if fn == nil {
		return "<nil>"
	}

	var sb strings.Builder
	sb.WriteString("(")

	for i, arg := range fn.Args {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(arg.String())
	}

	sb.WriteString(")")

	switch len(fn.Returns) {
	case 0:
	case 1:
		sb.WriteString(" -> ")
		sb.WriteString(fn.Returns[0].String())
	default:
		sb.WriteString(" -> (")
		for i, r := range fn.Returns {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(r.String())
		}
		sb.WriteString(")")
	}

	return sb.String()
}

func (init *Initializer) String() string {
	if init == nil {
		return "<nil>"
	}

	parts := make([]string, len(init.Args))
	for i, arg := range init.Args {
		parts[i] = arg.String()
	}

	return fmt.Sprintf("init(%s)", strings.Join(parts, ", "))
}

func (r *Record) String() string {
	if r == nil {
		return "<nil>"
	}

	var sb strings.Builder
	sb.WriteString("record {\n")
	for _, f := range r.Fields {
		sb.WriteString("  ")
		sb.WriteString(f.String())
		sb.WriteString("\n")
	}
	sb.WriteString("}")

	return sb.String()
}

func (s *Struct) String() string {
	if s == nil {
		return "<nil>"
	}

	var sb strings.Builder
	sb.WriteString("struct")
	if len(s.Impl) > 0 {
		sb.WriteString(" is ")
		for i, v := range s.Impl {
			if i != 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(v.String())
		}
	}
	sb.WriteString(" {\n")
	for _, f := range s.Fields {
		sb.WriteString("  ")
		sb.WriteString(f.String())
		sb.WriteString("\n")
	}
	for _, init := range s.Inits {
		sb.WriteString("  ")
		sb.WriteString(init.String())
		sb.WriteString("\n")
	}
	sb.WriteString("}")

	return sb.String()
}

func (iface *Interface) String() string {
	if iface == nil {
		return "<nil>"
	}

	var sb strings.Builder
	sb.WriteString("interface")

	if len(iface.Generics) > 0 {
		sb.WriteString("<")
		for i, g := range iface.Generics {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(g.String())
		}
		sb.WriteString(">")
	}

	sb.WriteString(" {\n")

	for _, c := range iface.Constants {
		sb.WriteString("  const ")
		sb.WriteString(c.String())
		sb.WriteString("\n")
	}
	for _, m := range iface.Methods {
		sb.WriteString("  fn ")
		sb.WriteString(m.String())
		sb.WriteString("\n")
	}

	sb.WriteString("}")

	return sb.String()
}

func (ec *EnumCase) String() string {
	if ec == nil {
		return "<nil>"
	}

	if len(ec.Args) == 0 {
		return ec.Name
	}

	parts := make([]string, len(ec.Args))
	for i, arg := range ec.Args {
		parts[i] = arg.String()
	}

	return fmt.Sprintf("%s(%s)", ec.Name, strings.Join(parts, ", "))
}

func (e *Enum) String() string {
	if e == nil {
		return "<nil>"
	}

	var sb strings.Builder
	sb.WriteString("enum {\n")
	for _, c := range e.Cases {
		sb.WriteString("  ")
		sb.WriteString(c.String())
		sb.WriteString("\n")
	}
	sb.WriteString("}")

	return sb.String()
}

func (vc *VariantCase) String() string {
	if vc == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s(%s)", vc.Name, vc.Type.String())
}

func (v *Variant) String() string {
	if v == nil {
		return "<nil>"
	}

	var sb strings.Builder
	sb.WriteString("variant {\n")
	for _, c := range v.Cases {
		sb.WriteString("  ")
		sb.WriteString(c.String())
		sb.WriteString("\n")
	}
	sb.WriteString("}")

	return sb.String()
}

func (ti *TypeDecl) String() string {
	if ti == nil {
		return "<nil>"
	}

	var name string
	if ti.setName {
		name = ti.name + " = "
	}

	var body string
	switch ti.Kind {
	case TypeRecord:
		body = ti.Record.String()
	case TypeStruct:
		body = ti.Struct.String()
	case TypeInterface:
		body = ti.Interface.String()
	case TypeEnum:
		body = ti.Enum.String()
	case TypeVariant:
		body = ti.Variant.String()
	case TypeFunction:
		body = ti.Function.String()
	default:
		body = "<unknown kind>"
	}

	return name + body
}
