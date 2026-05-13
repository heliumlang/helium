package vm

import (
	"fmt"
	"os"

	"github.com/Nykenik24/oxy/internal/frontend/lexer"
	"github.com/Nykenik24/oxy/internal/frontend/parser"
	"github.com/Nykenik24/oxy/internal/oxyerr"
	"github.com/Nykenik24/oxy/internal/util"
)

type TypeStructure int

type PlainType struct {
	Name string
}

type ArrayType struct {
	Base any
}

type FuncType struct {
	Args    []any
	Returns *any
}

type Qualifier int

const (
	Const Qualifier = iota
)

type Variable struct {
	Type       any
	Qualifiers util.Set[Qualifier]
}

type Context struct {
	Variables map[string]*Variable
	Filename  *string
}

type OxyDebug int

const (
	DebugNone OxyDebug = iota
	DebugAll
	DebugTokens
	DebugAST
)

type OxyVM struct {
	lexer  lexer.Lexer
	parser *parser.Parser
	source string
	Ctx    *Context
	debug  OxyDebug
}

func New() *OxyVM {
	return &OxyVM{
		lexer: lexer.New(),
		Ctx: &Context{
			Variables: make(map[string]*Variable),
			Filename:  nil,
		},
		debug: DebugNone,
	}
}

func (vm *OxyVM) LoadString(src string) {
	vm.source = src
}

func (vm *OxyVM) LoadFile(path string) *oxyerr.Error {
	content, err := os.ReadFile(path)
	if err != nil {
		return oxyerr.New(err.Error(), oxyerr.EmptyTrace())
	}

	vm.source = string(content)
	vm.Ctx.Filename = &path
	return nil
}

func (vm *OxyVM) SetDebug(d OxyDebug) {
	vm.debug = d
}

func (vm *OxyVM) debugAll() bool {
	return vm.debug == DebugAll
}

func (vm *OxyVM) Run() *oxyerr.Error {
	var filename string
	if vm.Ctx.Filename != nil {
		filename = *vm.Ctx.Filename
	} else {
		filename = "<src>"
	}

	vm.lexer.SetFilename(filename)
	tokens, err := vm.lexer.Lex(vm.source)

	if err != nil {
		return err
	}

	if vm.debugAll() || vm.debug == DebugTokens {
		for _, tk := range tokens {
			fmt.Println(tk)
		}
		fmt.Println()
	}

	vm.parser = parser.New(filename, tokens)
	nodes := vm.parser.Parse()

	if vm.debugAll() || vm.debug == DebugAST {
		for _, node := range nodes {
			fmt.Println(node.String())
		}
	}

	return nil
}
