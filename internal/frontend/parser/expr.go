/*
 * Expression parsing
 */

package parser

import (
	"fmt"

	"github.com/Nykenik24/oxy/internal/frontend/lexer"
)

func (p *Parser) parseExpr() Node {
	ti := p.enterRule("parse expression")
	defer p.traceRm(ti)
	return p.parseTernary()
}

func (p *Parser) parseTernary() Node {
	ti := p.enterRule("parse ternary")
	defer p.traceRm(ti)
	cond := p.parseAssign()
	if p.match(lexer.OpQuestion) {
		p.advance()
		_then := p.parseExpr()
		p.mustSkip(lexer.PunctColon)
		_else := p.parseTernary()
		return TernaryExpr{Cond: cond, Then: _then, Else: _else}
	}
	return cond
}

func (p *Parser) parseAssign() Node {
	ti := p.enterRule("parse assignment")
	defer p.traceRm(ti)
	left := p.parseOr()
	for p.oneOf(
		lexer.OpAssign,
		lexer.OpAssignAdd,
		lexer.OpAssignSub,
		lexer.OpAssignMul,
		lexer.OpAssignDiv,
		lexer.OpAssignMod,
	) {
		op := p.get(0).Lexeme()
		p.advance()
		right := p.parseOr()
		left = BinaryExpr{Left: left, Right: right, Op: op}
	}
	return left
}

func (p *Parser) parseOr() Node {
	ti := p.enterRule("parse or")
	defer p.traceRm(ti)
	left := p.parseAnd()
	for p.match(lexer.KeywordOr) {
		p.advance()
		right := p.parseAnd()
		left = BinaryExpr{Left: left, Right: right, Op: "or"}
	}
	return left
}

func (p *Parser) parseAnd() Node {
	ti := p.enterRule("parse and")
	defer p.traceRm(ti)
	left := p.parseEquality()
	for p.match(lexer.KeywordAnd) {
		p.advance()
		right := p.parseEquality()
		left = BinaryExpr{Left: left, Right: right, Op: "and"}
	}
	return left
}

func (p *Parser) parseEquality() Node {
	ti := p.enterRule("parse equality")
	defer p.traceRm(ti)
	left := p.parseRelational()
	for p.oneOf(lexer.OpEquals, lexer.OpNotEquals) {
		op := p.get(0).Lexeme()
		p.advance()
		right := p.parseRelational()
		left = BinaryExpr{Left: left, Right: right, Op: op}
	}
	return left
}

func (p *Parser) parseRelational() Node {
	ti := p.enterRule("parse relational")
	defer p.traceRm(ti)
	left := p.parseCoalesce()
	for p.oneOf(lexer.OpGreater, lexer.OpSmaller, lexer.OpEqGreater, lexer.OpEqSmaller) {
		op := p.get(0).Lexeme()
		p.advance()
		right := p.parseCoalesce()
		left = BinaryExpr{Left: left, Right: right, Op: op}
	}
	return left
}

func (p *Parser) parseCoalesce() Node {
	ti := p.enterRule("parse coalesce")
	defer p.traceRm(ti)
	left := p.parseAdditive()
	for p.match(lexer.OpFallback) {
		p.advance()
		var (
			right Node = nil
			block []Node
		)
		if p.match(lexer.PunctLBrace) {
			block = p.parseBlock()
		} else {
			right = p.parseAdditive()
		}
		left = CoalesceExpr{Left: left, Right: right, Block: block}
	}
	return left
}

func (p *Parser) parseAdditive() Node {
	ti := p.enterRule("parse additive")
	defer p.traceRm(ti)
	left := p.parseMultiplicative()
	for p.oneOf(lexer.OpAdd, lexer.OpSub) {
		op := p.get(0).Lexeme()
		p.advance()
		right := p.parseMultiplicative()
		left = BinaryExpr{Left: left, Right: right, Op: op}
	}
	return left
}

func (p *Parser) parseMultiplicative() Node {
	ti := p.enterRule("parse multiplicative")
	defer p.traceRm(ti)
	left := p.parseUnary()
	for p.oneOf(lexer.OpMul, lexer.OpDiv, lexer.OpMod) {
		op := p.get(0).Lexeme()
		p.advance()
		right := p.parseUnary()
		left = BinaryExpr{Left: left, Right: right, Op: op}
	}
	return left
}

func (p *Parser) parseUnary() Node {
	ti := p.enterRule("parse unary")
	defer p.traceRm(ti)
	if p.oneOf(
		lexer.OpSub,
		lexer.OpExclamation,
		lexer.OpIncrement,
		lexer.OpDecrement,
	) {
		op := p.get(0).Lexeme()
		p.advance()
		return UnaryExpr{Operand: p.parseUnary(), Op: op}
	}
	return p.parseCatch()
}

func (p *Parser) parseCatch() Node {
	ti := p.enterRule("parse catch")
	defer p.traceRm(ti)
	operand := p.parsePostfix()
	if p.match(lexer.KeywordCatch) {
		p.advance()
		name := p.mustRead(lexer.Ident)
		body := p.parseBlock()
		return CatchExpr{Operand: operand, ErrIdent: name, Body: body}
	}
	return operand
}

func (p *Parser) parsePostfix() Node {
	ti := p.enterRule("parse postfix")
	defer p.traceRm(ti)
	node := p.parsePrimary()
	for {
		t := p.get(0)
		switch t.Kind() {
		case lexer.OpExclamation:
			p.advance()
			node = ForceUnwrap{Operand: node}

		case lexer.OpQuestion:
			next := p.get(1).Kind()
			if next == lexer.PunctPeriod || next == lexer.PunctLBracket {
				p.advance()
				node = OptionalChain{Operand: node}
			} else {
				return node
			}

		case lexer.PunctLBracket:
			p.advance()
			index := p.parseExpr()
			p.mustSkip(lexer.PunctRBracket)
			node = IndexExpr{Object: node, Index: index}

		case lexer.PunctLParen:
			p.advance()
			args := list(p, lexer.PunctComma, lexer.PunctRParen, p.parseArg)
			node = CallExpr{Callee: node, Args: args}

		case lexer.PunctPeriod:
			p.advance()
			field := p.mustRead(lexer.Ident)
			if p.match(lexer.PunctLParen) {
				p.advance()
				args := list(p, lexer.PunctComma, lexer.PunctRParen, p.parseArg)
				node = MethodCall{Object: node, Method: field, Args: args}
			} else {
				node = FieldAccess{Object: node, Field: field}
			}

		default:
			return node
		}
	}
}

func (p *Parser) parsePrimary() Node {
	ti := p.enterRule("parse primary")
	defer p.traceRm(ti)
	t := p.get(0)
	switch t.Kind() {
	case lexer.KeywordNew:
		return p.parseNew()

	case lexer.KeywordFn:
		return p.parseClosure()

	case lexer.OpAt:
		p.advance()
		return SelfField{Field: p.mustRead(lexer.Ident)}

	case lexer.PunctLParen:
		p.advance()
		inner := p.parseExpr()
		p.mustSkip(lexer.PunctRParen)
		return GroupExpr{Inner: inner}

	case lexer.PunctLBracket:
		return p.parseArrayLit()

	case lexer.PunctLBrace:
		return p.parseMapLit()

	case lexer.KeywordTrue:
		p.advance()
		return BoolLit{Value: true}

	case lexer.KeywordFalse:
		p.advance()
		return BoolLit{Value: false}

	case lexer.KeywordNone:
		p.advance()
		return NoneLit{}

	case lexer.Integer:
		raw := p.get(0).Lexeme()
		p.advance()
		var v int64
		fmt.Sscanf(raw, "%d", &v)
		return IntLit{Value: v}

	case lexer.Float:
		raw := p.get(0).Lexeme()
		p.advance()
		var v float64
		fmt.Sscanf(raw, "%f", &v)
		return FloatLit{Value: v}

	case lexer.String:
		raw := p.get(0).Lexeme()
		p.advance()
		return StringLit{Value: raw}

	case lexer.Char:
		raw := p.get(0).Lexeme()
		p.advance()
		return CharLit{Value: raw[0]}

	case lexer.Ident:
		name := p.get(0).Lexeme()
		p.advance()
		return Ident{Name: name}

	default:
		p.error(fmt.Sprintf("unexpected \x1b[33m%s\x1b[0m in expression", t.Kind()), t.Pos())
		return nil
	}
}

func (p *Parser) parseNew() Node {
	ti := p.enterRule("parse new expression")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordNew)
	typeName := p.mustRead(lexer.Ident)
	var typeArgs []Node
	if p.match(lexer.OpSmaller) {
		typeArgs = p.parseTypeArgs()
	}
	p.mustSkip(lexer.PunctLParen)
	args := list(p, lexer.PunctComma, lexer.PunctRParen, p.parseArg)
	return NewExpr{TypeName: typeName, TypeArgs: typeArgs, Args: args}
}

func (p *Parser) parseClosure() Node {
	ti := p.enterRule("parse closure")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordFn)
	p.mustSkip(lexer.PunctLParen)
	params := list(p, lexer.PunctComma, lexer.PunctRParen, p.parseDeclArg)
	var returns []Node
	if !p.match(lexer.PunctLBrace) {
		returns = append(returns, p.parseType())
		for p.match(lexer.PunctComma) {
			p.advance()
			returns = append(returns, p.parseType())
		}
	}
	body := p.parseBlock()
	return ClosureExpr{Params: params, Returns: returns, Body: body}
}

func (p *Parser) parseArrayLit() Node {
	ti := p.enterRule("parse array literal")
	defer p.traceRm(ti)
	p.mustSkip(lexer.PunctLBracket)
	elements := list(p, lexer.PunctComma, lexer.PunctRBracket, p.parseExpr)
	return ArrayLit{Elements: elements}
}

func (p *Parser) parseMapLit() Node {
	ti := p.enterRule("parse map literal")
	defer p.traceRm(ti)
	p.mustSkip(lexer.PunctLBrace)
	elements := list(p, lexer.PunctComma, lexer.PunctRBrace, p.parseMapPair)
	return MapLit{Elements: elements}
}

func (p *Parser) parseMapPair() Node {
	key := p.parseExpr()
	p.mustSkip(lexer.PunctColon)
	value := p.parseExpr()
	return MapPair{
		Key:   key,
		Value: value,
	}
}

func (p *Parser) parseArg() Arg {
	ti := p.enterRule("parse argument")
	defer p.traceRm(ti)

	for p.match(lexer.NewLine) {
		p.advance()
	}

	var arg Arg
	if p.match(lexer.Ident) && p.get(1).Kind() == lexer.PunctColon {
		name := p.mustRead(lexer.Ident)
		arg.Name = &name
		p.mustSkip(lexer.PunctColon)
	}
	arg.Value = p.parseExpr()
	if p.match(lexer.PunctEllipsis) {
		p.advance()
		arg.Spread = true
	}

	for p.match(lexer.NewLine) {
		p.advance()
	}

	return arg
}
