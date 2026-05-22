package types

func (b *BaseType) Equal(other *BaseType) bool {
	if b == nil && other == nil {
		return true
	}
	if b == nil || other == nil {
		return false
	}

	if b.Optional != other.Optional || b.Throwable != other.Throwable {
		return false
	}

	if b.IsArray != other.IsArray || b.IsMap != other.IsMap || b.IsClosure != other.IsClosure {
		return false
	}

	switch {
	case b.IsArray:
		return b.Array.Compare(other.Array)
	case b.IsMap:
		return b.Map.Compare(other.Map)
	case b.IsClosure:
		return b.Closure.Compare(other.Closure)
	default:
		if !CompareStringPtr(b.Name, other.Name) {
			return false
		}
		return CompareTypeSlice(b.TypeArgs, other.TypeArgs)
	}
}

func (a *Array) Compare(other *Array) bool {
	if a == nil && other == nil {
		return true
	}
	if a == nil || other == nil {
		return false
	}
	return a.Nesting == other.Nesting && a.Value.Equal(other.Value)
}

func (m *Map) Compare(other *Map) bool {
	if m == nil && other == nil {
		return true
	}
	if m == nil || other == nil {
		return false
	}
	return m.Key.Equal(other.Key) && m.Value.Equal(other.Value)
}

func (c *Closure) Compare(other *Closure) bool {
	if c == nil && other == nil {
		return true
	}
	if c == nil || other == nil {
		return false
	}
	if len(c.Args) != len(other.Args) || len(c.Returns) != len(other.Returns) {
		return false
	}
	for i, arg := range c.Args {
		if arg.Name != other.Args[i].Name {
			return false
		}
		if !arg.Type.Equal(other.Args[i].Type) {
			return false
		}
	}
	for i, ret := range c.Returns {
		if !ret.Equal(other.Returns[i]) {
			return false
		}
	}
	return true
}

func CompareStringPtr(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func CompareTypeSlice(a, b []*BaseType) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Equal(b[i]) {
			return false
		}
	}
	return true
}
