/*
 * Statement parsing
 */

package parser

import (
	"fmt"

	"github.com/heliumlang/helium/internal/frontend/lexer"
)

func (p *Parser) parseStatement() Node {
	ti := p.enterRule("parse statement")
	defer p.traceRm(ti)
	for p.match(lexer.NewLine) {
		p.advance()
	}
	t := p.get(0)
	switch t.Kind() {
	case lexer.KeywordUse:
		return p.parseUse()
	case lexer.KeywordExtern:
		return p.parseExtern()
	case lexer.Ident, lexer.KeywordConst, lexer.OpIncrement, lexer.OpDecrement, lexer.OpAt:
		if p.isDeclAssign() {
			return p.parseVarDecl()
		}
		return p.parseExprStmt()
	case lexer.KeywordReturn:
		p.advance()
		var exprs []Node
		for !p.match(lexer.NewLine) && !p.match(lexer.PunctRBrace) && !p.isEOF() {
			exprs = append(exprs, p.parseExpr())
			if !p.match(lexer.PunctComma) {
				break
			}
			p.advance()
		}
		if p.match(lexer.NewLine) {
			p.advance()
		}
		return Return{Exprs: exprs}
	case lexer.KeywordIf:
		return p.parseIfStmt()
	case lexer.KeywordFor:
		return p.parseForStmt()
	case lexer.KeywordDo:
		return p.parseDoStmt()
	case lexer.KeywordSwitch:
		return p.parseSwitchStmt()
	case lexer.KeywordRaise:
		p.advance()
		expr := p.parseExpr()
		return Raise{Expr: expr}
	default:
		p.error(
			fmt.Sprintf("expected statement, got \x1b[33m%s\x1b[0m", t.Kind()),
			t.Pos(),
		)
		return nil
	}
}

func (p *Parser) isDeclAssign() bool {
	i := 0
	for {
		tok := p.get(i)
		if tok == nil {
			return false
		}
		switch tok.Kind() {
		case lexer.Ident, lexer.KeywordConst:
			i++
		case lexer.PunctComma:
			i++
		case lexer.OpAssignNew:
			return true
		default:
			return false
		}
	}
}

func (p *Parser) parseExprStmt() Node {
	ti := p.enterRule("parse expression statement")
	defer p.traceRm(ti)
	expr := p.parseExpr()
	if p.match(lexer.NewLine) {
		p.advance()
	}
	return ExprStmt{Expr: expr}
}

func (p *Parser) parseIfStmt() Node {
	ti := p.enterRule("parse if statement")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordIf)
	cond := p.parseExpr()
	body := p.parseBlock()
	var elifs []Elif
	var els *[]Node
	for {
		for p.match(lexer.NewLine) {
			p.advance()
		}
		if !p.match(lexer.KeywordElse) {
			break
		}
		p.advance()
		if p.match(lexer.KeywordIf) {
			p.advance()
			elifCond := p.parseExpr()
			elifBody := p.parseBlock()
			elifs = append(elifs, Elif{Cond: elifCond, Body: elifBody})
		} else {
			elseBody := p.parseBlock()
			els = &elseBody
			break
		}
	}
	return IfStmt{Cond: cond, Body: body, Elifs: elifs, Else: els}
}

func (p *Parser) parseBlock() []Node {
	ti := p.enterRule("parse block")
	defer p.traceRm(ti)
	p.mustSkip(lexer.PunctLBrace)
	var body []Node
	for {
		for p.match(lexer.NewLine) {
			p.advance()
		}
		if p.match(lexer.PunctRBrace) || p.isEOF() {
			break
		}
		body = append(body, p.parseStatement())
	}

	p.mustSkip(lexer.PunctRBrace)
	return body
}

func (p *Parser) parseForStmt() Node {
	ti := p.enterRule("parse for statement")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordFor)
	idents := list(p, lexer.PunctComma, lexer.KeywordIn, func() string {
		return p.mustRead(lexer.Ident)
	})
	iter := p.parseExpr()
	body := p.parseBlock()
	return For{Idents: idents, Iter: iter, Body: body}
}

func (p *Parser) parseDoStmt() Node {
	ti := p.enterRule("parse do-while statement")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordDo)
	body := p.parseBlock()
	p.mustSkip(lexer.KeywordWhile)
	cond := p.parseExpr()
	return Do{Cond: cond, Body: body}
}

func (p *Parser) parseSwitchStmt() Node {
	ti := p.enterRule("parse switch statement")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordSwitch)
	operand := p.parseExpr()
	p.mustSkip(lexer.PunctLBrace)
	var (
		cases []SwitchCase
		def   *SwitchResult
	)
	for {
		for p.match(lexer.NewLine) {
			p.advance()
		}
		if p.match(lexer.PunctRBrace) || p.isEOF() {
			break
		}
		if p.match(lexer.KeywordDefault) {
			p.advance()
			p.mustSkip(lexer.OpArrow)
			result := p.parseSwitchResult()
			def = &result
		} else {
			cases = append(cases, p.parseSwitchCase())
		}
	}
	p.mustSkip(lexer.PunctRBrace)
	var defVal SwitchResult
	if def != nil {
		defVal = *def
	}
	return Switch{Operand: operand, Cases: cases, Default: defVal}
}

func (p *Parser) parseSwitchCase() SwitchCase {
	ti := p.enterRule("parse switch case")
	defer p.traceRm(ti)

	expr := p.parseExpr()
	var params []string
	if p.match(lexer.PunctPipe) {
		p.advance()
		params = list(p, lexer.PunctComma, lexer.PunctPipe, func() string {
			return p.mustRead(lexer.Ident)
		})
	}

	p.mustSkip(lexer.OpArrow)
	return SwitchCase{
		Pattern: SwitchPattern{
			Expr:   expr,
			Params: params,
		},
		Result: p.parseSwitchResult(),
	}
}

func (p *Parser) parseSwitchResult() SwitchResult {
	ti := p.enterRule("parse switch result")
	defer p.traceRm(ti)
	if p.match(lexer.PunctLBrace) {
		return SwitchResult{Block: p.parseBlock()}
	}
	expr := p.parseExpr()
	if p.match(lexer.NewLine) {
		p.advance()
	}
	return SwitchResult{Expr: &expr}
}
