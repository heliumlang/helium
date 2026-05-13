package lexer

import (
	"errors"
	"fmt"

	"github.com/Nykenik24/oxy/internal/oxyerr"
)

type Lexer interface {
	Lex(input string) ([]*Token, *oxyerr.Error)
	SetFilename(fname string)
}

type lexer struct {
	tokens   []*Token
	input    string
	n        int
	i        int
	filename string
}

func New() Lexer {
	return &lexer{}
}

func (l *lexer) SetFilename(fname string) {
	l.filename = fname
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

	lexeme = append(lexeme, l.input[l.i])
	l.i++

	for l.inbounds() && (isAlpha(l.curr()) || isDigit(l.curr()) || l.curr() == '_') {
		lexeme = append(lexeme, l.curr())
		l.i++
	}

	if ok, k := isKeyword(lexeme); ok {
		kind = k
	}
	return NewToken(lexeme, kind), nil
}

func (l *lexer) lexDigit() (*Token, error) {
	var (
		lexeme []byte
		kind   TokenKind = Integer
	)

	for l.inbounds() && isDigit(l.curr()) {
		lexeme = append(lexeme, l.curr())
		l.i++
	}
	if l.inbounds() && l.curr() == '.' {
		kind = Float
		lexeme = append(lexeme, l.curr())
		l.i++
		for l.inbounds() && isDigit(l.curr()) {
			lexeme = append(lexeme, l.curr())
			l.i++
		}
	}

	return NewToken(lexeme, kind), nil
}

func (l *lexer) lexString() (*Token, error) {
	var lexeme []byte
	if l.curr() == '"' {
		l.i++
	}
	for l.inbounds() && l.curr() != '"' && l.curr() != '\n' {
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

func (l *lexer) Lex(input string) ([]*Token, *oxyerr.Error) {
	l.input = input
	l.n = len(input)
	l.i = 0
	line, col := 1, 1

	for l.inbounds() {
		starti := l.i
		char := input[l.i]

		if char == '#' {
			l.i++
			for l.curr() != '\n' {
				l.i++
			}
			continue
		}

		var (
			token *Token = NewToken([]byte(""), None)
			err   error  = nil
		)

		startCol := col
		if isWhitespace(char) {
			if char == '\n' || char == '\r' {
				line++
				col = 1
				token = NewToken([]byte{char}, NewLine)
			} else {
				col++
			}
			l.i++
			if token.kind == None {
				continue
			}
		} else if isAlpha(char) || char == '_' {
			token, err = l.lexIdent()
		} else if isDigit(char) {
			token, err = l.lexDigit()
		} else if char == '"' {
			token, err = l.lexString()
		} else if char == '\'' {
			token, err = l.lexChar()
		} else {
			matched := false
			for _, length := range []int{3, 2, 1} {
				if l.i+length > l.n {
					continue
				}
				slice := l.input[l.i : l.i+length]
				if ok, k, _ := isOp(slice); ok {
					token = NewToken([]byte(slice), k)
					l.i += length
					matched = true
					break
				}
				if ok, k, _ := isPunct(slice); ok {
					token = NewToken([]byte(slice), k)
					l.i += length
					matched = true
					break
				}
			}
			if !matched {
				l.i++
			}
		}
		col += l.i - starti

		if err != nil {
			return nil, oxyerr.New(err.Error(), oxyerr.EmptyTrace()).SetType("lex")
		}

		token.line = line
		token.col = startCol
		l.tokens = append(l.tokens, token)
	}

	eof := NewToken([]byte(""), EOF)
	eof.line = line
	eof.col = col - 1
	l.tokens = append(l.tokens, eof)

	return l.tokens, nil
}
