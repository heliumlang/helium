/*
 * Polo's parser
 */

package parser

import (
	"os"
	"time"

	"github.com/Nykenik24/helium/internal/frontend/lexer"
	"github.com/Nykenik24/helium/internal/heliumerr"
)

type Parser struct {
	tokens   []*lexer.Token     // list of tokens
	index    int                // the index in the token list
	filename string             // the filename of the source
	trace    []*heliumerr.Trace // the trace stack
}

func New(file string, tokens []*lexer.Token) *Parser {
	return &Parser{
		tokens:   tokens,
		index:    0,
		filename: file,
	}
}

// get tokens[index + n]
func (p *Parser) get(n int) *lexer.Token {
	if !p.inbounds(n) {
		return nil
	}
	return p.tokens[p.index+n]
}

// consume token
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

// add a rule to the trace stack
func (p *Parser) enterRule(name string) int {
	entry := &heliumerr.Trace{
		Name:    name,
		Entered: time.Now(),
		File:    "<helium>",
	}
	p.trace = append(p.trace, entry)
	return len(p.trace) - 1
}

// remove a rule from the trace stack with it's
// trace index
func (p *Parser) traceRm(i int) {
	if i < 0 || i >= len(p.trace) {
		return
	}
	p.trace = p.trace[:i]
}

// error & panic
func (p *Parser) error(msg string, pos lexer.Position) {
	err := heliumerr.New(msg, p.trace)
	err.SetPos(pos.Line, pos.Col).SetFilename(p.filename).SetType("parse").Print()

	panic(parseError{})
}

type parseError struct{}

// check if the current token is nil/EOF
func (p *Parser) isEOF() bool {
	t := p.get(0)
	return t == nil || t.Kind() == lexer.EOF
}

func (p *Parser) Parse() Node {
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
	return prog
}
