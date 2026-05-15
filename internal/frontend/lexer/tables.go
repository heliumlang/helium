/*
 * Tables for the lexer.
 *
 * The lexer uses these to map certain keywords/operators/punctuations
 * to the corresponding token kind.
 */

package lexer

type word struct {
	str  string
	kind TokenKind
}

// keyword table
var reserved = []word{
	w("fn", KeywordFn),
	w("mod", KeywordModule),
	w("struct", KeywordStruct),
	w("record", KeywordRecord),
	w("interface", KeywordInterface),
	w("is", KeywordIs),
	w("new", KeywordNew),
	w("return", KeywordReturn),
	w("const", KeywordConst),
	w("and", KeywordAnd),
	w("or", KeywordOr),
	w("catch", KeywordCatch),
	w("true", KeywordTrue),
	w("false", KeywordFalse),
	w("none", KeywordNone),
	w("if", KeywordIf),
	w("else", KeywordElse),
	w("for", KeywordFor),
	w("in", KeywordIn),
	w("do", KeywordDo),
	w("while", KeywordWhile),
	w("switch", KeywordSwitch),
	w("default", KeywordDefault),
	w("raise", KeywordRaise),
	w("init", KeywordInit),
	w("enum", KeywordEnum),
	w("variant", KeywordVariant),
	w("alias", KeywordAlias),
	w("use", KeywordUse),
	w("from", KeywordFrom),
	w("export", KeywordExport),
	w("extern", KeywordExtern),
	w("comp", KeywordCompile),
}

// operator table
var operators = []word{
	w("+", OpAdd),
	w("-", OpSub),
	w("*", OpMul),
	w("/", OpDiv),
	w("%", OpMod),

	w(":=", OpAssignNew),
	w("=", OpAssign),
	w("+=", OpAssignAdd),
	w("-=", OpAssignSub),
	w("*=", OpAssignMul),
	w("/=", OpAssignDiv),
	w("%=", OpAssignMod),

	w("==", OpEquals),
	w("!=", OpNotEquals),
	w(">", OpGreater),
	w("<", OpSmaller),
	w(">=", OpEqGreater),
	w("<=", OpEqSmaller),

	w("++", OpIncrement),
	w("--", OpDecrement),

	w("??", OpFallback),
	w("!", OpExclamation),
	w("?", OpQuestion),
	w("@", OpAt),
	w("=>", OpArrow),
}

// punctuation table
var punct = []word{
	w(".", PunctPeriod),
	w(",", PunctComma),
	w(":", PunctColon),
	w(";", PunctSemicolon),
	w("(", PunctLParen),
	w(")", PunctRParen),
	w("[", PunctLBracket),
	w("]", PunctRBracket),
	w("{", PunctLBrace),
	w("}", PunctRBrace),
	w("|", PunctPipe),
	w("...", PunctEllipsis),
}

func w(str string, kind TokenKind) word {
	return word{
		str:  str,
		kind: kind,
	}
}

// DO NOT CHANGE
var initializedTokenNames = false

var tokenNames = map[TokenKind]string{
	None:     "Null token",
	EOF:      "EOF",
	Ident:    "Identifier",
	Reserved: "Reserved",
	Integer:  "Integer",
	Float:    "Float",
	String:   "String",
	Char:     "Char",
	NewLine:  "New line",
	Shortcut: "Shortcut",
}
