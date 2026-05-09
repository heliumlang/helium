package parser

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Nykenik24/oxy/internal/lexer"
)

type traceEntry struct {
	name    string
	entered time.Time
	depth   int
}

type Parser struct {
	tokens   []*lexer.Token
	index    int
	filename string
	trace    []traceEntry
}

func New(file string, tokens []*lexer.Token) *Parser {
	return &Parser{
		tokens:   tokens,
		index:    0,
		filename: file,
	}
}

func (p *Parser) get(n int) *lexer.Token {
	if !p.inbounds(n) {
		return nil
	}
	return p.tokens[p.index+n]
}

func (p *Parser) advance() *lexer.Token {
	p.index++
	if !p.inbounds(0) {
		p.error(
			"unexpected end of input",
			p.get(-1).Pos(),
		)
	}
	return p.get(0)
}

func (p *Parser) enterRule(name string) int {
	entry := traceEntry{
		name:    name,
		entered: time.Now(),
		depth:   len(p.trace),
	}
	p.trace = append(p.trace, entry)
	return len(p.trace) - 1
}

func (p *Parser) traceRm(i int) {
	if i < 0 || i >= len(p.trace) {
		return
	}
	p.trace = p.trace[:i]
}

func (p *Parser) error(msg string, pos lexer.Position) {
	snapshot := make([]traceEntry, len(p.trace))
	copy(snapshot, p.trace)

	fmt.Printf("\x1b[91mparsing error\x1b[0m in file \x1b[32m%s\x1b[0m at \x1b[35m%d:%d\x1b[0m:\n",
		p.filename, pos.Line, pos.Col)
	fmt.Printf("\x1b[90m└──\x1b[0m %s\n", msg)
	fmt.Println()
	fmt.Println("\x1b[34mTrace stack:\x1b[0m")

	if len(snapshot) == 0 {
		fmt.Println("  \x1b[90m(empty — no parse rules active)\x1b[0m")
	}

	for i, entry := range snapshot {
		indent := strings.Repeat("  ", entry.depth)
		elapsed := time.Since(entry.entered)

		connector := ""
		if entry.depth > 0 {
			connector = "└──"
		}

		fmt.Printf("\x1b[90m%s%s\x1b[0m [\x1b[33m%d\x1b[0m] %s \x1b[90m(%s)\x1b[0m\n",
			indent, connector, i, entry.name, elapsed.Round(time.Microsecond))
	}

	if t := p.get(0); t != nil {
		fmt.Printf("\n\x1b[90mcurrent token:\x1b[0m \x1b[32m'%s'\x1b[0m \x1b[90m(kind=%v)\x1b[0m\n",
			t.Lexeme(), t.Kind())
	}

	panic(parseError{})
}

type parseError struct{}

func (p *Parser) isEOF() bool {
	t := p.get(0)
	return t == nil || t.Kind() == lexer.EOF
}

func (p *Parser) Parse() []Node {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(parseError); !ok {
				panic(r)
			}
			os.Exit(1)
		}
	}()

	prog := &Program{}
	prog.Items = append(prog.Items, p.parseModule())
	for !p.isEOF() {
		for p.match(lexer.Newline) {
			p.advance()
		}
		if p.isEOF() {
			break
		}
		prog.Items = append(prog.Items, p.parseDecl())
	}
	return []Node{prog}
}
