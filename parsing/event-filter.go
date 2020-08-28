package parsing

import (
	"bytes"
	"errors"
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
	Unknown TokenKind = iota
	Or
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
	bytes := []byte(input)
	var tokens []Token

	for len(bytes) > 0 {
		token, err := advanceToken(&bytes)
		if err != nil {
			return NewTokenStream(), err
		}
		tokens = append(tokens, token)
	}

	return NewTokenStream(tokens...), nil
}

func advanceToken(b *[]byte) (Token, error) {
	// skip whitespaces
	*b = bytes.TrimSpace(*b)

	switch {
	case bytes.HasPrefix(*b, []byte("||")):
		*b = (*b)[2:]
		return Token{Or, ""}, nil

	case bytes.HasPrefix(*b, []byte("&&")):
		*b = (*b)[2:]
		return Token{And, ""}, nil

	case bytes.HasPrefix(*b, []byte("==")):
		*b = (*b)[2:]
		return Token{Eq, ""}, nil

	case bytes.HasPrefix(*b, []byte("!=")):
		*b = (*b)[2:]
		return Token{Neq, ""}, nil

	case (*b)[0] == '(':
		*b = (*b)[1:]
		return Token{LParen, ""}, nil

	case (*b)[0] == ')':
		*b = (*b)[1:]
		return Token{RParen, ""}, nil
	}

	return Token{Unknown, ""}, errors.New("Unknown token")
}
