package parser

import (
	"fmt"
	"slices"

	"github.com/Nykenik24/oxy/internal/frontend/lexer"
)

func (p *Parser) inbounds(n int) bool {
	i := p.index + n
	return i >= 0 && i < len(p.tokens)
}

func (p *Parser) match(kind lexer.TokenKind) bool {
	t := p.get(0)
	return t != nil && t.Kind() == kind
}

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

func (p *Parser) mustSkip(kind lexer.TokenKind) *lexer.Token {
	p.must(kind)
	return p.advance()
}

func (p *Parser) mustRead(kind lexer.TokenKind) string {
	lexeme := p.must(kind).Lexeme()
	p.advance()
	return lexeme
}

func (p *Parser) oneOf(kinds ...lexer.TokenKind) bool {
	return slices.Contains(kinds, p.get(0).Kind())
}

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

func block[T Node](p *Parser, open, close lexer.TokenKind, fn func() T) []T {
	var body []T

	p.mustSkip(open)
	for !p.match(close) {
		body = append(body, fn())
	}
	p.mustSkip(close)

	return body
}
