package parser

import (
	"fmt"

	"github.com/Nykenik24/oxy/internal/lexer"
)

func (p *Parser) parseType() Node {
	ti := p.enterRule("parse type")
	defer p.traceRm(ti)

	t := p.get(0)
	switch t.Kind() {
	case lexer.KeywordFn:
		return p.parseFunctionType()

	case lexer.Ident:
		name := p.mustRead(lexer.Ident)

		var typeArgs []Node
		if p.match(lexer.OpSmaller) {
			typeArgs = p.parseTypeArgs()
		}

		base := BaseType{Typename: name, TypeArgs: typeArgs}

		if p.match(lexer.PunctLBracket) {
			return p.parseArrayType(base)
		}

		if p.match(lexer.OpQuestion) {
			base.Optional = true
			p.advance()
		}
		if p.match(lexer.OpExclamation) {
			base.Throwable = true
			p.advance()
		}

		return base

	default:
		p.error(fmt.Sprintf(
			"expected type, got \x1b[33m%s\x1b[0m \x1b[34m%s\x1b[0m",
			t.Kind(), t.Lexeme(),
		), t.Pos())
		return nil
	}
}

func (p *Parser) parseFunctionType() Node {
	ti := p.enterRule("parse function type")
	defer p.traceRm(ti)

	p.mustSkip(lexer.KeywordFn)
	p.mustSkip(lexer.PunctLParen)
	args := list(p, lexer.PunctComma, lexer.PunctRParen, p.parseType)
	var returns Node
	if p.oneOf(lexer.Ident, lexer.KeywordFn) {
		returns = p.parseType()
	}
	return FunctionType{Args: args, Returns: returns}
}

func (p *Parser) parseArrayType(base Node) Node {
	ti := p.enterRule("parse array tpe")
	defer p.traceRm(ti)

	node := base
	for p.match(lexer.PunctLBracket) {
		p.mustSkip(lexer.PunctLBracket)
		p.mustSkip(lexer.PunctRBracket)
		node = ArrayType{Values: node}
	}
	return node
}

func (p *Parser) parseTypeArgs() []Node {
	ti := p.enterRule("parse type args")
	defer p.traceRm(ti)

	p.mustSkip(lexer.OpSmaller)
	return list(p, lexer.PunctComma, lexer.OpGreater, p.parseType)
}
