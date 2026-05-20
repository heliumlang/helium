package check

import (
	"errors"
	"fmt"
)

type TypeTable struct {
	types   []*TypeInfo
	aliases map[string]*Base
}

func NewTypeTable() *TypeTable {
	return &TypeTable{}
}

func (tt *TypeTable) Lookup(name string) (int, bool) {
	for i, t := range tt.types {
		if t.name == name {
			return i, true
		}
	}

	return -1, false
}

func (tt *TypeTable) Get(name string) *TypeInfo {
	for _, t := range tt.types {
		if t.name == name {
			return t
		}
	}

	return nil
}

func (tt *TypeTable) Register(t *TypeInfo) error {
	if !t.HasName() {
		return errors.New("type doesn't have name")
	}

	if _, ok := tt.Lookup(t.GetName()); ok {
		return fmt.Errorf("type %s already defined", t.GetName())
	}

	tt.types = append(tt.types, t)
	return nil
}

func (tt *TypeTable) Overwrite(t *TypeInfo) error {
	if !t.HasName() {
		return errors.New("type doesn't have name")
	}

	i, ok := tt.Lookup(t.GetName())
	if !ok {
		return fmt.Errorf("couldn't overwrite: type %s not defined", t.GetName())
	}

	tt.types[i] = t
	return nil
}

func (tt *TypeTable) Alias(name string, target *Base) error {
	if _, defined := tt.aliases[name]; defined {
		return fmt.Errorf("alias %s already defined", name)
	}

	tt.aliases[name] = target
	return nil
}

func (tt *TypeTable) GetAlias(name string) (*Base, error) {
	if _, defined := tt.aliases[name]; !defined {
		return nil, fmt.Errorf("alias %s not found", name)
	}

	return tt.aliases[name], nil
}

func (tt *TypeTable) All() []*TypeInfo {
	return tt.types
}
