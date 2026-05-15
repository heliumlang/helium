/*
 * Parsing helpers
 */

package parser

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Nykenik24/polo/internal/frontend/lexer"
)

// check if index >= 0 && index < len(tokens)
func (p *Parser) inbounds(n int) bool {
	i := p.index + n
	return i >= 0 && i < len(p.tokens)
}

// check if current matches kind
func (p *Parser) match(kind lexer.TokenKind) bool {
	t := p.get(0)
	return t != nil && t.Kind() == kind
}

// panic if current doesn't match kind
func (p *Parser) must(kind lexer.TokenKind) *lexer.Token {
	t := p.get(0)
	if t == nil || t.Kind() != kind {
		gotKind := "<nil>"
		if t != nil {
			gotKind = fmt.Sprintf("%s", t.Kind())
		}
		pos := p.get(-1).Pos()
		if t != nil {
			pos = t.Pos()
		}
		p.error(fmt.Sprintf("expected \x1b[33m%s\x1b[0m, got \x1b[33m%s\x1b[0m",
			kind, gotKind), pos)
	}
	return t
}

// panic if current doesn't match kind,
// if it does skip the token
func (p *Parser) mustSkip(kind lexer.TokenKind) *lexer.Token {
	p.must(kind)
	return p.advance()
}

// panic if current doesn't match kind,
// if it does return the lexeme
func (p *Parser) mustRead(kind lexer.TokenKind) string {
	lexeme := p.must(kind).Lexeme()
	p.advance()
	return lexeme
}

// check if current matches one of kinds
func (p *Parser) oneOf(kinds ...lexer.TokenKind) bool {
	return slices.Contains(kinds, p.get(0).Kind())
}

// panic if current doesn't match one of kinds
func (p *Parser) mustOneOf(kinds ...lexer.TokenKind) *lexer.Token {
	if p.oneOf(kinds...) {
		t := p.get(0)
		p.advance()
		return t
	}
	var parts []string
	for _, k := range kinds {
		parts = append(parts, k.String())
	}
	p.error(fmt.Sprintf("expected one of %s", strings.Join(parts, ", ")), p.get(0).Pos())
	return nil
}

// create a list
func list[T any](p *Parser, sep, end lexer.TokenKind, fn func() T) []T {
	var items []T
	for !p.match(end) {
		for p.match(lexer.NewLine) {
			p.advance()
		}
		if p.match(end) {
			break
		}
		items = append(items, fn())
		for p.match(lexer.NewLine) {
			p.advance()
		}
		if !p.match(sep) {
			break
		}
		p.advance()
		for p.match(lexer.NewLine) {
			p.advance()
		}
	}
	p.mustSkip(end)
	return items
}
