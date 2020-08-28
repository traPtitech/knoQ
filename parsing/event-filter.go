package parsing

import (
	"regexp"
)

/*---------------------------------------------------------------------------*/

type TokenStream struct {
	tokens []Token
	pos    int
}

func NewTokenStream(tokens ...Token) TokenStream {
	return TokenStream{tokens, 0}
}

func (ts *TokenStream) HasNext() bool {
	return ts.pos > len(ts.tokens)
}

func (ts *TokenStream) Next() Token {
	ts.pos++
	return ts.tokens[ts.pos]
}

type Token struct {
	Kind  TokenKind
	Value string // for Attr, UUID
}

type TokenKind int

const (
	Or TokenKind = iota
	And
	Eq
	Neq
	LParen
	RParen
	Attr
	UUID
)

/*---------------------------------------------------------------------------*/

var (
	SupportedAttributes = []string{"user", "group", "tag", "event"}
	ReAttrOrUUIDLike    = regexp.MustCompile(`^[a-z0-9\-:{}]*`)
)

func CheckAttrOrUUIDLike(lexeme string) TokenKind {
	for _, attr := range SupportedAttributes {
		if attr == lexeme {
			return Attr
		}
	}
	return UUID
}

/*---------------------------------------------------------------------------*/

func Lex(input string) (TokenStream, error) {
	return NewTokenStream(), nil
}
