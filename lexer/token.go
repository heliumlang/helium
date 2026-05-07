package lexer

import "fmt"

type TokenKind int

const (
	None TokenKind = iota
	EOF
	Ident
	Reserved
	Digit
	String
	Char
	Symbol
	Punct
)

// fill later
var reserved = []string{
	"if",
}

// fill later
var symbols = []string{
	"+", "-", "*", "/", "=",
}

// fill later
var punct = []string{
	".", ":", ";", "(", ")", "[", "]", "{", "}",
}

var kindToString = map[TokenKind]string{
	None:     "Null token",
	EOF:      "EOF",
	Ident:    "Identifier",
	Reserved: "Reserved",
	Digit:    "Digit",
	String:   "String",
	Char:     "Char",
	Symbol:   "Symbol",
	Punct:    "Punctuation",
}

func (k TokenKind) String() string {
	str, ok := kindToString[k]
	if !ok {
		return "Unknown"
	}
	return str
}

type Token struct {
	lexeme []byte
	kind   TokenKind
}

func NewToken(lexeme []byte, kind TokenKind) *Token {
	return &Token{
		lexeme: lexeme,
		kind:   kind,
	}
}

// for later (parser)
func (t *Token) Match(s string) bool {
	return string(t.lexeme) == s
}

func (t *Token) String() string {
	return fmt.Sprintf("Lexeme(\x1b[32m'%s'\x1b[0m), Kind(\x1b[35m%s\x1b[0m)", t.lexeme, t.kind.String())
}
