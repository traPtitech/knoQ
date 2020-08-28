package parsing

import (
	"bytes"
	"errors"
	"regexp"

	"github.com/gofrs/uuid"
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
	return ts.pos < len(ts.tokens)-1
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
	ReAttrOrUUIDLike    = regexp.MustCompile(`^[a-z0-9\-:{}]+`)
)

func checkAttrOrUUIDLike(lexeme string) TokenKind {
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

	loc := ReAttrOrUUIDLike.FindIndex(*b)

	var token Token
	switch {
	case loc != nil:
		match := string((*b)[:loc[1]])
		kind := checkAttrOrUUIDLike(match)
		if kind == Attr {
			token = Token{kind, match}
		} else if uuid, err := uuid.FromString(match); err == nil {
			token = Token{UUID, uuid.String()}
		} else {
			return Token{Unknown, ""}, err
		}
		*b = (*b)[loc[1]:]

	case bytes.HasPrefix(*b, []byte("||")):
		token = Token{Or, ""}
		*b = (*b)[2:]

	case bytes.HasPrefix(*b, []byte("&&")):
		token = Token{And, ""}
		*b = (*b)[2:]

	case bytes.HasPrefix(*b, []byte("==")):
		token = Token{Eq, ""}
		*b = (*b)[2:]

	case bytes.HasPrefix(*b, []byte("!=")):
		token = Token{Neq, ""}
		*b = (*b)[2:]

	case (*b)[0] == '(':
		token = Token{LParen, ""}
		*b = (*b)[1:]

	case (*b)[0] == ')':
		token = Token{RParen, ""}
		*b = (*b)[1:]

	default:
		return Token{Unknown, ""}, errors.New("Unknown token")
	}

	return token, nil
}
