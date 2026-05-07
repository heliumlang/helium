package lexer

import (
	"errors"
	"fmt"
	"slices"
)

type Lexer interface {
	Lex(input string) ([]*Token, error)
}

type lexer struct {
	tokens []*Token
	input  string
	n      int
	i      int
}

func New() Lexer {
	return &lexer{}
}

func (l *lexer) inbounds() bool {
	return l.i < l.n
}

func (l *lexer) curr() byte {
	return l.input[l.i]
}

func (l *lexer) lexIdent() (*Token, error) {
	var lexeme []byte
	kind := Ident

	for l.inbounds() && (isAlpha(l.curr()) || l.curr() == '_') {
		lexeme = append(lexeme, l.curr())
		l.i++
	}

	if slices.Contains(reserved, string(lexeme)) {
		kind = Reserved
	}
	return NewToken(lexeme, kind), nil
}

func (l *lexer) lexDigit() (*Token, error) {
	var lexeme []byte
	for l.inbounds() && isDigit(l.curr()) {
		lexeme = append(lexeme, l.curr())
		l.i++
	}
	if l.inbounds() && l.curr() == '.' {
		lexeme = append(lexeme, l.curr())
		l.i++
		for l.inbounds() && isDigit(l.curr()) {
			lexeme = append(lexeme, l.curr())
			l.i++
		}
	}

	return NewToken(lexeme, Digit), nil
}

func (l *lexer) lexString() (*Token, error) {
	var lexeme []byte
	if l.curr() == '"' {
		l.i++
	}
	for l.inbounds() && l.curr() != '"' {
		lexeme = append(lexeme, l.curr())
		l.i++
	}

	if !l.inbounds() || l.curr() != '"' {
		return nil, errors.New("unterminated string")
	}

	l.i++

	return NewToken(lexeme, String), nil
}

func (l *lexer) lexChar() (*Token, error) {
	var lexeme []byte
	if l.curr() == '\'' {
		l.i++
	}
	if !l.inbounds() {
		return nil, fmt.Errorf("out of bounds when lexing character")
	}
	lexeme = append(lexeme, l.curr())
	l.i++

	if !l.inbounds() || l.curr() != '\'' {
		return nil, errors.New("unterminated char")
	}

	l.i++

	return NewToken(lexeme, Char), nil
}

func (l *lexer) Lex(input string) ([]*Token, error) {
	l.input = input
	l.n = len(input)
	l.i = 0

	for l.inbounds() {
		char := input[l.i]
		if isWhitespace(char) {
			l.i++
			continue
		}

		var (
			token *Token = NewToken([]byte(""), None)
			err   error  = nil
		)

		if isAlpha(char) || char == '_' {
			token, err = l.lexIdent()
		} else if isDigit(char) {
			token, err = l.lexDigit()
		} else if char == '"' {
			token, err = l.lexString()
		} else if char == '\'' {
			token, err = l.lexChar()
		} else if isSym(char) {
			token = NewToken([]byte{char}, Symbol)
			l.i++
		} else if isPunct(char) {
			token = NewToken([]byte{char}, Punct)
			l.i++
		} else {
			l.i++
		}

		if err != nil {
			return nil, err
		}

		if len(token.lexeme) > 0 {
			l.tokens = append(l.tokens, token)
		}
	}
	l.tokens = append(l.tokens, NewToken([]byte(""), EOF))
	return l.tokens, nil
}
