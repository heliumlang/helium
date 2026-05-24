package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/heliumlang/helium/internal/compiler"
	"github.com/heliumlang/helium/internal/frontend/lexer"
	"github.com/heliumlang/helium/internal/frontend/parser"
	"github.com/heliumlang/helium/internal/heliumerr"
)

type debug int

const (
	debugTokens debug = 1 << iota
	debugAST
	debugAll = debugTokens | debugAST
)

func (d debug) has(flag debug) bool {
	return d&flag != 0
}

func main() {
	start := time.Now()

	if err := run(); err != nil {
		err.SetFilename(flag.Args()[0]).Print()
		os.Exit(1)
	}

	elapsed := time.Since(start)
	fmt.Printf("Took \x1b[32m%s\x1b[0m\n", elapsed)

}

func run() *heliumerr.Error {
	var dbgFlag = flag.Int("debug", int(debugAll), "debug: 1=tokens, 2=ast, 3=all, 0=none")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: helium [flags] <filename>")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		fmt.Println("missing required argument: filename")
		os.Exit(1)
	}

	dbg := debug(*dbgFlag)
	path := args[0]

	source, err := os.ReadFile(path)
	if err != nil {
		return heliumerr.Wrap(err)
	}

	lex := lexer.New()
	lex.SetFilename(path)

	var tokens []*lexer.Token

	tokens, err = lex.Lex(string(source))
	if err != nil {
		return heliumerr.Wrap(err).SetType("lexical")
	}

	if dbg.has(debugTokens) {
		for _, tk := range tokens {
			fmt.Println(tk)
		}
		fmt.Println()
	}

	parse := parser.New(path, tokens)

	ast := parse.Parse()

	if dbg.has(debugAST) {
		fmt.Println(ast)
	}

	c := compiler.NewCompiler()
	err = c.Compile(ast)
	if err != nil {
		heliumerr.Wrap(err).SetType("compile").SetFilename(path).Print()
		os.Exit(1)
	}

	for _, t := range c.GetTypes() {
		fmt.Println()
		fmt.Println(t)
	}

	fmt.Println(c.Dissasemble())
	fmt.Println()
	fmt.Println("=== full bytecode ===")
	fmt.Println(c.StringSerialize())

	return nil
}
