/*
 * Statement parsing
 */

package parser

import (
	"fmt"

	"github.com/heliumlang/helium/internal/frontend/lexer"
)

func (p *Parser) parseModule() Node {
	ti := p.enterRule("parse module")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordModule)
	name := p.mustRead(lexer.Ident)
	p.mustSkip(lexer.NewLine)
	return &Module{Name: name}
}

func (p *Parser) parseUse() Node {
	ti := p.enterRule("parse use")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordUse)

	if p.match(lexer.Shortcut) {
		return &Use{
			Kind: UseNamespace,
			Path: p.parsePath(),
		}
	}

	var (
		kind    UseKind
		members []string
	)

	if p.match(lexer.OpMul) {
		p.advance()
		kind = UseWildcard
	} else {
		kind = UseMembers
		members = append(members, p.mustRead(lexer.Ident))
		for p.match(lexer.PunctComma) {
			p.advance()
			if !p.match(lexer.Ident) {
				break
			}
			members = append(members, p.mustRead(lexer.Ident))
		}
	}

	p.mustSkip(lexer.KeywordFrom)

	return &Use{
		Kind:    kind,
		Path:    p.parsePath(),
		Members: members,
	}
}

func (p *Parser) parsePath() []string {
	var path []string
	path = append(path, p.mustOneOf(lexer.Ident, lexer.Shortcut).Lexeme())
	for p.match(lexer.OpDiv) {
		p.advance()
		path = append(path, p.mustOneOf(lexer.Ident, lexer.Shortcut).Lexeme())
	}
	return path
}

func (p *Parser) parseExtern() Node {
	ti := p.enterRule("parse extern")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordExtern)
	if p.match(lexer.PunctLBrace) {
		p.advance()
		members := list(p, lexer.PunctComma, lexer.PunctRBrace, func() string {
			return p.mustRead(lexer.Ident)
		})
		return Extern{Members: members}
	} else {
		return Extern{Members: []string{p.mustRead(lexer.Ident)}}
	}
}

func (p *Parser) parseStmt() Node {
	ti := p.enterRule("parse statement")
	defer p.traceRm(ti)
	for p.match(lexer.NewLine) {
		p.advance()
	}
	t := p.get(0)
	switch t.Kind() {
	case lexer.KeywordNoop:
		p.advance()
		return Noop{}
	case lexer.KeywordUse:
		return p.parseUse()
	case lexer.KeywordExtern:
		return p.parseExtern()
	case lexer.Ident, lexer.KeywordConst, lexer.KeywordCompile, lexer.OpIncrement, lexer.OpDecrement, lexer.OpAt:
		if p.isDeclAssign() {
			return p.parseVarDecl()
		}
		return p.parseExprStmt()
	case lexer.KeywordReturn:
		p.advance()
		exprs := p.exprList()
		return Return{Exprs: exprs}
	case lexer.KeywordDefer:
		p.advance()
		exprs := p.exprList()
		return Defer{Exprs: exprs}
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
		case lexer.Ident, lexer.KeywordConst, lexer.KeywordCompile:
			i++
		case lexer.PunctComma:
			i++
		case lexer.PunctColon:
			for p.inbounds(i) && p.get(i).Kind() != lexer.OpAssignNew {
				i++
				if p.index+i > len(p.tokens) {
					p.error("unterminated variable declaration", tok.Pos())
				}
			}
		case lexer.OpAssignNew:
			return true
		default:
			return false
		}
	}
}

func (p *Parser) exprList() []Node {
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
	return exprs
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
	for p.match(lexer.NewLine) {
		p.advance()
	}
	if p.match(lexer.KeywordNoop) {
		p.advance()
		if !p.match(lexer.PunctRBrace) {
			p.error("expected end of block after no-op", p.get(0).Pos())
			return nil
		}
		p.advance()
		return []Node{Noop{}}
	}
	var body []Node
	for {
		for p.match(lexer.NewLine) {
			p.advance()
		}
		if p.match(lexer.PunctRBrace) || p.isEOF() {
			break
		}
		body = append(body, p.parseStmt())
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
