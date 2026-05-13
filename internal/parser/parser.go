package parser

import (
	"os"
	"time"

	oxy_error "github.com/Nykenik24/oxy/internal/error"
	"github.com/Nykenik24/oxy/internal/lexer"
)

type Parser struct {
	tokens   []*lexer.Token
	index    int
	filename string
	trace    []*oxy_error.Trace
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
	entry := &oxy_error.Trace{
		Name:    name,
		Entered: time.Now(),
		File:    "<oxy>",
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
	err := oxy_error.New(msg, p.trace)
	err.SetPos(pos.Line, pos.Col).SetFilename(p.filename).SetType("parse").Print()

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
		for p.match(lexer.NewLine) {
			p.advance()
		}
		if p.isEOF() {
			break
		}
		prog.Items = append(prog.Items, p.parseDecl())
	}
	return []Node{prog}
}
