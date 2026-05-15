/*
 * Declaration parsing
 */

package parser

import (
	"fmt"
	"slices"

	"github.com/Nykenik24/polo/internal/frontend/lexer"
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

	var (
		from     *From = nil
		wildcard       = false
		members  []string
	)

	if p.match(lexer.OpMul) {
		wildcard = true
		p.advance()
	} else {
		members = append(members, p.mustRead(lexer.Ident))
		for p.match(lexer.PunctComma) {
			p.advance()
			if p.match(lexer.Ident) {
				members = append(members, p.get(0).Lexeme())
				p.advance()
			} else {
				break
			}
		}
	}

	if p.match(lexer.KeywordFrom) {
		p.advance()
		var path []*lexer.Token
		path = append(path, p.mustOneOf(lexer.Ident, lexer.Shortcut))
		for p.match(lexer.OpDiv) {
			p.advance()
			path = append(path, p.mustOneOf(lexer.Ident, lexer.Shortcut))
		}
		from = &From{Path: path}
	}

	return &Use{
		From:     from,
		Members:  members,
		Wildcard: wildcard,
	}
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

func (p *Parser) parseDecl() Node {
	ti := p.enterRule("parse declaration")
	defer p.traceRm(ti)
	var annotations []Annotation
	for p.match(lexer.OpAt) {
		annotations = append(annotations, p.parseAnnotation())
		for p.match(lexer.NewLine) {
			p.advance()
		}
	}
	t := p.get(0)
	switch t.Kind() {
	case lexer.KeywordUse:
		return p.parseUse()
	case lexer.KeywordExtern:
		return p.parseExtern()
	case lexer.Ident:
		return p.parseVarDecl()
	case lexer.KeywordFn, lexer.KeywordExport:
		return p.parseFuncWithAnnotations(annotations)
	case lexer.KeywordStruct:
		return p.parseStruct()
	case lexer.KeywordRecord:
		return p.parseRecord()
	case lexer.KeywordInterface:
		return p.parseInterface()
	case lexer.KeywordEnum:
		return p.parseEnum()
	case lexer.KeywordVariant:
		return p.parseVariant()
	case lexer.KeywordAlias:
		return p.parseAlias()
	case lexer.KeywordConst:
		if p.isDeclAssign() {
			return p.parseVarDecl()
		}
		return p.parseConst()
	default:
		p.error(fmt.Sprintf("expected declaration, got \x1b[33m%s\x1b[0m", t.Kind()), t.Pos())
		return nil
	}
}

func (p *Parser) parseAnnotation() Annotation {
	ti := p.enterRule("parse annotation")
	defer p.traceRm(ti)
	p.mustSkip(lexer.OpAt)
	name := p.mustRead(lexer.Ident)
	var value string
	if p.match(lexer.PunctLParen) {
		p.advance()
		value = p.mustRead(lexer.String)
		p.mustSkip(lexer.PunctRParen)
	}
	return Annotation{Name: name, Value: value}
}

func (p *Parser) parseVarDecl() Node {
	ti := p.enterRule("parse variable declaration")
	defer p.traceRm(ti)
	_const := false
	if p.match(lexer.KeywordConst) {
		_const = true
		p.advance()
	}
	idents := list(p, lexer.PunctComma, lexer.OpAssignNew, func() string {
		return p.mustRead(lexer.Ident)
	})
	var exprs []Node
	for {
		exprs = append(exprs, p.parseExpr())
		if !p.match(lexer.PunctComma) {
			break
		}
		p.advance()
	}
	if p.match(lexer.NewLine) {
		p.advance()
	}
	return VarDecl{Idents: idents, Exprs: exprs, Const: _const}
}

func (p *Parser) parseFuncWithAnnotations(annotations []Annotation) Node {
	ti := p.enterRule("parse function declaration")
	defer p.traceRm(ti)
	public := false
	if p.match(lexer.KeywordExport) {
		p.advance()
		public = true
	}
	p.mustSkip(lexer.KeywordFn)
	var (
		recv    *Receiver = nil
		returns []Node
	)
	if p.match(lexer.PunctLBracket) {
		p.advance()
		recv = &Receiver{}
		recv.Name = p.mustRead(lexer.Ident)
		recv.Type = p.parseType()
		p.mustSkip(lexer.PunctRBracket)
	}
	name := p.mustRead(lexer.Ident)
	var typeArgs []Node
	if p.match(lexer.OpSmaller) {
		typeArgs = p.parseTypeArgs()
	}
	p.mustSkip(lexer.PunctLParen)
	args := list(p, lexer.PunctComma, lexer.PunctRParen, p.parseDeclArg)
	if !p.match(lexer.PunctLBrace) {
		returns = list(p, lexer.PunctComma, lexer.PunctLBrace, p.parseType)
		p.index--
	}
	body := p.parseBlock()
	return &FunctionDecl{
		Public:      public,
		Name:        name,
		Args:        args,
		TypeArgs:    typeArgs,
		Body:        body,
		Recv:        recv,
		Returns:     returns,
		Annotations: annotations,
	}
}

func (p *Parser) parseRecord() Node {
	ti := p.enterRule("parse record declaration")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordRecord)
	name := p.mustRead(lexer.Ident)
	var generics []Node
	if p.match(lexer.OpSmaller) {
		generics = p.parseTypeArgs()
	}
	fields := p.parseRecordBody()
	return Record{
		Name:     name,
		Generics: generics,
		Fields:   fields,
	}
}

func (p *Parser) parseRecordBody() []Node {
	ti := p.enterRule("parse record body")
	defer p.traceRm(ti)
	p.mustSkip(lexer.PunctLBrace)
	var fields []Node
	for !p.match(lexer.PunctRBrace) {
		for p.match(lexer.NewLine) {
			p.advance()
		}
		if p.match(lexer.PunctRBrace) {
			break
		}
		fields = append(fields, p.parseStructField())
		p.mustSkip(lexer.NewLine)
	}
	p.mustSkip(lexer.PunctRBrace)
	return fields
}

func (p *Parser) parseInterface() Node {
	ti := p.enterRule("parse interface declaration")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordInterface)
	name := p.mustRead(lexer.Ident)
	var generics []Node
	if p.match(lexer.OpSmaller) {
		generics = p.parseTypeArgs()
	}
	p.mustSkip(lexer.PunctLBrace)
	var members []Node
	for !p.match(lexer.PunctRBrace) {
		for p.match(lexer.NewLine) {
			p.advance()
		}
		if p.match(lexer.PunctRBrace) {
			break
		}
		switch p.get(0).Kind() {
		case lexer.KeywordFn:
			members = append(members, p.parseFnSig())
		case lexer.KeywordConst:
			members = append(members, p.parseConst())
		default:
			p.error(fmt.Sprintf("expected interface member, got \x1b[33m%s\x1b[0m", p.get(0).Kind()), p.get(0).Pos())
		}
	}
	p.mustSkip(lexer.PunctRBrace)
	return Interface{
		Name:     name,
		Generics: generics,
		Members:  members,
	}
}

func (p *Parser) parseFnSig() Node {
	ti := p.enterRule("parse function signature")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordFn)
	name := p.mustRead(lexer.Ident)
	p.mustSkip(lexer.PunctLParen)
	args := list(p, lexer.PunctComma, lexer.PunctRParen, p.parseDeclArg)
	var returns []Node
	if !p.match(lexer.NewLine) && !p.match(lexer.PunctRBrace) {
		returns = append(returns, p.parseType())
		for p.match(lexer.PunctComma) {
			p.advance()
			returns = append(returns, p.parseType())
		}
	}
	return FnSig{Name: name, Args: args, Returns: returns}
}

func (p *Parser) parseEnum() Node {
	ti := p.enterRule("parse enum declaration")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordEnum)
	name := p.mustRead(lexer.Ident)
	p.mustSkip(lexer.PunctLBrace)
	var variants []EnumVariant
	for !p.match(lexer.PunctRBrace) {
		for p.match(lexer.NewLine) {
			p.advance()
		}
		if p.match(lexer.PunctRBrace) {
			break
		}
		variants = append(variants, p.parseEnumVariant())
	}
	p.mustSkip(lexer.PunctRBrace)
	return Enum{Name: name, Variants: variants}
}

func (p *Parser) parseEnumVariant() EnumVariant {
	ti := p.enterRule("parse enum variant")
	defer p.traceRm(ti)
	name := p.mustRead(lexer.Ident)
	var params []DeclArg
	if p.match(lexer.PunctLParen) {
		p.advance()
		params = list(p, lexer.PunctComma, lexer.PunctRParen, p.parseDeclArg)
	}
	for p.match(lexer.NewLine) {
		p.advance()
	}
	return EnumVariant{Name: name, Params: params}
}

func (p *Parser) parseVariant() Node {
	ti := p.enterRule("parse variant declaration")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordVariant)
	name := p.mustRead(lexer.Ident)
	p.mustSkip(lexer.PunctLBrace)
	var fields []VariantField
	for !p.match(lexer.PunctRBrace) {
		for p.match(lexer.NewLine) {
			p.advance()
		}
		if p.match(lexer.PunctRBrace) {
			break
		}
		fields = append(fields, p.parseVariantField())
	}
	p.mustSkip(lexer.PunctRBrace)
	return Variant{Name: name, Fields: fields}
}

func (p *Parser) parseVariantField() VariantField {
	ti := p.enterRule("parse variant field")
	defer p.traceRm(ti)
	t := p.parseType()
	name := p.mustRead(lexer.Ident)
	for p.match(lexer.NewLine) {
		p.advance()
	}
	return VariantField{Type: t, Name: name}
}

func (p *Parser) parseAlias() Node {
	ti := p.enterRule("parse alias declaration")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordAlias)
	name := p.mustRead(lexer.Ident)
	p.mustSkip(lexer.OpAssign)
	t := p.parseType()
	return Alias{Name: name, Type: t}
}

func (p *Parser) parseConst() Node {
	ti := p.enterRule("parse const declaration")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordConst)
	t := p.parseType()
	name := p.mustRead(lexer.Ident)
	p.mustSkip(lexer.OpAssign)
	expr := p.parseExpr()
	return Const{Type: t, Name: name, Expr: expr}
}

func (p *Parser) parseStruct() Node {
	ti := p.enterRule("parse struct declaration")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordStruct)
	name := p.mustRead(lexer.Ident)
	var generics []Node
	if p.match(lexer.OpSmaller) {
		generics = p.parseTypeArgs()
	}
	var interfaces []Node
	if p.match(lexer.KeywordIs) {
		p.advance()
		interfaces = list(p, lexer.PunctComma, lexer.PunctLBrace, p.parseType)
		p.index--
	}
	fields, inits := p.parseStructBody()
	return Struct{
		Name:       name,
		Generics:   generics,
		Interfaces: interfaces,
		Fields:     fields,
		Inits:      inits,
	}
}

func (p *Parser) parseStructBody() ([]StructField, []Init) {
	ti := p.enterRule("parse struct body")
	defer p.traceRm(ti)
	p.mustSkip(lexer.PunctLBrace)
	var (
		fields []StructField
		inits  []Init
	)
	for !p.match(lexer.PunctRBrace) {
		for p.match(lexer.NewLine) {
			p.advance()
		}
		if p.match(lexer.PunctRBrace) {
			break
		}
		if p.match(lexer.KeywordInit) {
			inits = append(inits, p.parseInit())
			continue
		}
		fields = append(fields, p.parseStructField())
		p.mustSkip(lexer.NewLine)
	}
	p.mustSkip(lexer.PunctRBrace)
	return fields, inits
}

func (p *Parser) parseInit() Init {
	ti := p.enterRule("parse init declaration")
	defer p.traceRm(ti)
	p.mustSkip(lexer.KeywordInit)
	p.mustSkip(lexer.PunctLParen)
	params := list(p, lexer.PunctComma, lexer.PunctRParen, p.parseDeclArg)
	body := p.parseBlock()
	return Init{Params: params, Body: body}
}

func (p *Parser) parseStructField() StructField {
	ti := p.enterRule("parse struct field")
	defer p.traceRm(ti)
	var qualifiers []string
	for p.oneOf(
		lexer.KeywordConst,
		lexer.KeywordExport,
	) {
		qualif := p.get(0).Lexeme()
		if slices.Contains(qualifiers, qualif) {
			p.error("can't have more than one of the same qualifier in field", p.get(0).Pos())
		}
		qualifiers = append(qualifiers, qualif)
		p.advance()
	}
	_type := p.parseType()
	name := p.mustRead(lexer.Ident)
	return StructField{
		Qualifiers: qualifiers,
		Name:       name,
		Type:       _type,
	}
}

func (p *Parser) parseDeclArg() DeclArg {
	_type := p.parseType()
	variadic := p.match(lexer.PunctEllipsis)
	if variadic {
		p.advance()
	}
	name := p.mustRead(lexer.Ident)

	return DeclArg{
		Type:     _type,
		Name:     name,
		Variadic: variadic,
	}
}
