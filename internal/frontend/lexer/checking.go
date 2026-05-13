package lexer

import "slices"

func isWhitespace(c byte) bool {
	return slices.Contains([]byte{' ', '\n', '\t', '\r'}, c)
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isOp(s string) (bool, TokenKind, int) {
	for _, sym := range operators {
		if s == sym.str {
			return true, sym.kind, len(sym.str)
		}
	}

	return false, None, -1
}

func isPunct(s string) (bool, TokenKind, int) {
	for _, sym := range punct {
		if s == sym.str {
			return true, sym.kind, len(sym.str)
		}
	}

	return false, None, -1
}

func isKeyword(b []byte) (bool, TokenKind) {
	for _, kw := range reserved {
		if string(b) == kw.str {
			return true, kw.kind
		}
	}

	return false, None
}
