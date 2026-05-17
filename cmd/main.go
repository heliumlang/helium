package main

import (
	"fmt"
	"os"
	"time"

	"github.com/heliumlang/helium/internal/frontend/lexer"
	"github.com/heliumlang/helium/internal/frontend/parser"
	"github.com/heliumlang/helium/internal/heliumerr"
)

func puttime(elapsed time.Duration) {
	fmt.Printf("took \x1b[34m~%s\x1b[0m \x1b[90m(%s)\x1b[0m to run\n", elapsed.Round(time.Microsecond), elapsed)
}

func benchmark(fn func()) {
	start := time.Now()

	fn()

	elapsed := time.Since(start)
	puttime(elapsed)
}

func multibench(fn func(), iter int) {
	start := time.Now()

	for range iter {
		fn()
	}

	elapsed := time.Since(start)
	puttime(elapsed)
}

type debug int

const (
	debugAll debug = iota
	debugTokens
	debugAST
	debugCompiler
)

var dbg = debugCompiler

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: helium <filename>")
		os.Exit(1)
	}

	lex := lexer.New()
	path := os.Args[1]
	lex.SetFilename(path)

	source, readerr := os.ReadFile(path)
	if readerr != nil {
		heliumerr.New(readerr.Error(), heliumerr.EmptyTrace()).Print()
		os.Exit(1)
	}

	tokens, err := lex.Lex(string(source))

	if err != nil {
		err.Print()
		os.Exit(1)
	}

	if dbg == debugAll || dbg == debugTokens {
		for _, tk := range tokens {
			fmt.Println(tk)
		}
		fmt.Println()
	}

	parse := parser.New(path, tokens)
	ast := parse.Parse()

	if dbg == debugAll || dbg == debugAST {
		fmt.Println(ast)
	}
}
