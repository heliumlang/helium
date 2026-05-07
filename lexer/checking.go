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

func isSym(c byte) bool {
	return slices.Contains(symbols, string(c))
}

func isPunct(c byte) bool {
	return slices.Contains(punct, string(c))
}
