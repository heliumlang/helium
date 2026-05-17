/*
 * Token kinds and token structure
 */

package lexer

import (
	"fmt"
	"strings"
)

type TokenKind int

const (
	None TokenKind = iota
	EOF
	NewLine
	Ident
	Reserved
	Integer
	Float
	String
	Char
	Shortcut
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpAssign
	OpAssignNew
	OpEquals
	OpNotEquals
	OpAssignAdd
	OpAssignSub
	OpAssignMul
	OpAssignDiv
	OpAssignMod
	OpGreater
	OpSmaller
	OpEqGreater
	OpEqSmaller
	OpIncrement
	OpDecrement
	OpFallback
	OpExclamation
	OpQuestion
	OpAt
	OpArrow
	PunctPeriod
	PunctComma
	PunctColon
	PunctSemicolon
	PunctLParen
	PunctRParen
	PunctLBracket
	PunctRBracket
	PunctLBrace
	PunctRBrace
	PunctEllipsis
	PunctPipe
	KeywordFn
	KeywordModule
	KeywordUse
	KeywordFrom
	KeywordStruct
	KeywordRecord
	KeywordInterface
	KeywordIs
	KeywordNew
	KeywordReturn
	KeywordConst
	KeywordAnd
	KeywordOr
	KeywordCatch
	KeywordTrue
	KeywordFalse
	KeywordNone
	KeywordIf
	KeywordElse
	KeywordFor
	KeywordIn
	KeywordDo
	KeywordWhile
	KeywordSwitch
	KeywordDefault
	KeywordRaise
	KeywordInit
	KeywordEnum
	KeywordVariant
	KeywordAlias
	KeywordExport
	KeywordExtern
	KeywordCompile
	KeywordDefer
)

func initTokenNames() {
	for _, sym := range operators {
		tokenNames[sym.kind] = fmt.Sprintf("Operator %q", sym.str)
	}

	for _, sym := range punct {
		tokenNames[sym.kind] = fmt.Sprintf("Punctuation %q", sym.str)
	}

	for _, w := range reserved {
		tokenNames[w.kind] = fmt.Sprintf("Keyword %q", w.str)
	}
}

func (k TokenKind) String() string {
	if !initializedTokenNames {
		initTokenNames()
		initializedTokenNames = true
	}

	str, ok := tokenNames[k]
	if !ok {
		return "Unknown"
	}
	return str
}

type Token struct {
	lexeme            []byte    // the text of the token
	kind              TokenKind // the kind of the token
	line, col, offset int       // line, column and offset (end) of the token in the source
}

type Position struct{ Line, Col, Offset int }

func ZeroPos() Position {
	return Position{
		Line:   0,
		Col:    0,
		Offset: 0,
	}
}

func NewToken(lexeme []byte, kind TokenKind) *Token {
	return &Token{
		lexeme: lexeme,
		kind:   kind,
	}
}

// get text
func (t *Token) Lexeme() string {
	return string(t.lexeme)
}

// get kind
func (t *Token) Kind() TokenKind {
	return t.kind
}

func (t *Token) String() string {
	return strings.ReplaceAll(fmt.Sprintf("Lexeme(\x1b[32m'%s'\x1b[0m), Kind(\x1b[35m%s\x1b[0m)", t.lexeme, t.kind.String()), "\n", "\\n")
}

func (t *Token) Line() int {
	return t.line
}

func (t *Token) Col() int {
	return t.col
}

func (t *Token) Offset() int {
	return t.offset
}

func (t *Token) Pos() Position {
	if t == nil {
		return Position{}
	}
	return Position{Line: t.Line(), Col: t.Col(), Offset: t.Offset()}
}

// get an empty token
func ZeroToken() *Token {
	return NewToken([]byte(""), None)
}
