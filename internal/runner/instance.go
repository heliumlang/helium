package runner

import (
	"fmt"
	"os"
	"time"

	"github.com/Nykenik24/oxy/internal/lexer"
	"github.com/Nykenik24/oxy/internal/parser"
)

type Results struct {
	Tokens []*lexer.Token
}

type Instance struct {
	lexer  lexer.Lexer
	parser *parser.Parser
	res    *Results
}

func New() *Instance {
	return &Instance{
		lexer:  lexer.New(),
		parser: nil,
		res: &Results{
			Tokens: []*lexer.Token{lexer.ZeroToken()},
		},
	}
}

func (i *Instance) Benchmark(fn func()) {
	start := time.Now()
	defer (func() {
		elapsed := time.Since(start)
		fmt.Println()
		fmt.Printf("Took \x1b[34m%v\x1b[0m to run\n", elapsed)
	})()

	fn()
}

func (i *Instance) run(raw string, filename string) error {
	tokens, err := i.lexer.Lex(raw)
	i.res.Tokens = tokens

	if err != nil {
		return err
	}

	for _, tk := range tokens {
		fmt.Println(tk)
	}
	fmt.Println()

	i.parser = parser.New(filename, tokens)
	nodes := i.parser.Parse()

	for _, node := range nodes {
		fmt.Println(node.String())
	}

	return nil
}

func (i *Instance) RunString(input string) error {
	return i.run(input, "<raw string>")
}

func (i *Instance) RunFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return i.run(string(content), filename)
}

func (i *Instance) Results() *Results {
	return i.res
}
