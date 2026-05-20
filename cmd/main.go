package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/heliumlang/helium/internal/compiler/check"
	"github.com/heliumlang/helium/internal/frontend/lexer"
	"github.com/heliumlang/helium/internal/frontend/parser"
	"github.com/heliumlang/helium/internal/heliumerr"
)

type debug int

const (
	debugTokens debug = 1 << iota
	debugAST
	debugTypes
	debugAll = debugTokens | debugAST | debugTypes
)

func (d debug) has(flag debug) bool {
	return d&flag != 0
}

func puttime(label string, elapsed time.Duration) {
	fmt.Fprintf(os.Stderr, "%s took \x1b[34m~%s\x1b[0m \x1b[90m(%s)\x1b[0m\n",
		label, elapsed.Round(time.Microsecond), elapsed)
}

func multibench(label string, fn func(), iter int) {
	start := time.Now()
	for range iter {
		fn()
	}
	puttime(fmt.Sprintf("%s x%d", label, iter), time.Since(start))
}

func main() {
	if err := run(); err != nil {
		err.Print()
		os.Exit(1)
	}
}

func run() *heliumerr.Error {
	var (
		dbgFlag   = flag.Int("debug", int(debugAll), "debug: 1=tokens, 2=ast, 4=types, 7=all, 0=none")
		bench     = flag.Bool("bench", false, "benchmark lexing and parsing")
		benchIter = flag.Int("bench-iter", 100, "number of benchmark iterations (requires -bench)")
	)

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: helium [flags] <filename>")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		return heliumerr.New("missing required argument: filename", heliumerr.EmptyTrace())
	}

	dbg := debug(*dbgFlag)
	path := args[0]

	source, readerr := os.ReadFile(path)
	if readerr != nil {
		return heliumerr.New(readerr.Error(), heliumerr.EmptyTrace())
	}

	lex := lexer.New()
	lex.SetFilename(path)

	var tokens []*lexer.Token

	lexFn := func() {
		var lexErr *heliumerr.Error
		tokens, lexErr = lex.Lex(string(source))
		if lexErr != nil {
			lexErr.Print()
			os.Exit(1)
		}
	}

	if *bench {
		multibench("lex", lexFn, *benchIter)
	} else {
		lexFn()
	}

	if dbg.has(debugTokens) {
		for _, tk := range tokens {
			fmt.Println(tk)
		}
		fmt.Println()
	}

	parse := parser.New(path, tokens)

	var ast parser.Node

	parseFn := func() {
		ast = parse.Parse()
	}

	if *bench {
		multibench("parse", parseFn, *benchIter)
	} else {
		parseFn()
	}

	if dbg.has(debugAST) {
		fmt.Println(ast)
	}

	err, typeTable := check.Check(path, ast)
	if dbg.has(debugTypes) {
		fmt.Println()
		fmt.Printf("Found %d types\n", len(typeTable.All()))
		fmt.Println()
		for _, t := range typeTable.All() {
			fmt.Println(t)
		}
	}

	if err != nil {
		return err
	}

	return nil
}
