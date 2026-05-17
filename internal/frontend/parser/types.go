/*
 * Type parsing
 */

package parser

import (
	"fmt"

	"github.com/heliumlang/helium/internal/frontend/lexer"
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
		base.Optional, base.Throwable = p.checkTypeQualifs()

		if p.match(lexer.PunctLBracket) {
			return p.parseArrayType(base)
		} else if p.match(lexer.PunctLBrace) {
			return p.parseMapType(base)
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
	var returns []Node
	if p.oneOf(lexer.Ident, lexer.KeywordFn) {
		returns = append(returns, p.parseType())
		for p.match(lexer.PunctComma) {
			p.advance()
			returns = append(returns, p.parseType())
		}
	}
	return FunctionType{Args: args, Returns: returns}
}

func (p *Parser) parseArrayType(base Node) Node {
	ti := p.enterRule("parse array type")
	defer p.traceRm(ti)
	node := base
	if !p.match(lexer.PunctLBracket) {
		p.error("expected left bracket in array type", p.get(0).Pos())
	}
	for p.match(lexer.PunctLBracket) {
		p.mustSkip(lexer.PunctLBracket)
		p.mustSkip(lexer.PunctRBracket)
		node = ArrayType{Values: node}
	}
	array := node.(ArrayType)
	array.Optional, array.Throwable = p.checkTypeQualifs()
	return array
}

func (p *Parser) parseMapType(base Node) Node {
	ti := p.enterRule("parse map type")
	defer p.traceRm(ti)
	node := base
	if !p.match(lexer.PunctLBrace) {
		p.error("expected left brace in map type", p.get(0).Pos())
	}
	for p.match(lexer.PunctLBrace) {
		p.mustSkip(lexer.PunctLBrace)
		key := p.parseType()
		p.mustSkip(lexer.PunctRBrace)
		node = MapType{Key: key, Value: node}
	}
	maptype := node.(MapType)
	maptype.Optional, maptype.Throwable = p.checkTypeQualifs()
	return maptype
}

func (p *Parser) parseTypeArgs() []Node {
	ti := p.enterRule("parse type args")
	defer p.traceRm(ti)

	p.mustSkip(lexer.OpSmaller)
	return list(p, lexer.PunctComma, lexer.OpGreater, p.parseType)
}

func (p *Parser) checkTypeQualifs() (opt, throw bool) {
	if p.match(lexer.OpQuestion) {
		opt = true
		p.advance()
	}
	if p.match(lexer.OpExclamation) {
		throw = true
		p.advance()
	}
	return opt, throw
}
