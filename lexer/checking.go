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
