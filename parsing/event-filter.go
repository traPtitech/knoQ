package parsing

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"

	"github.com/gofrs/uuid"
)

/*---------------------------------------------------------------------------*/

type TokenStream struct {
	tokens []Token
	pos    int
}

func NewTokenStream(tokens ...Token) TokenStream {
	tokens = append(tokens, Token{EOF, ""})
	return TokenStream{tokens, 0}
}

func (ts *TokenStream) HasNext() bool {
	return ts.Peek().Kind != EOF
}

func (ts *TokenStream) Next() Token {
	ts.pos++
	return ts.tokens[ts.pos]
}

func (ts *TokenStream) Peek() Token {
	return ts.tokens[ts.pos]
}

func (ts *TokenStream) Restore() {
	ts.pos = 0
}

type Token struct {
	Kind  tokenKind
	Value string // for Attr, UUID
}

type tokenKind int

const (
	Unknown tokenKind = iota
	Or
	And
	Eq
	Neq
	LParen
	RParen
	Attr
	UUID
	EOF
)

func (k tokenKind) String() string {
	switch k {
	case Unknown:
		return "unknown"
	case Or:
		return "||"
	case And:
		return "&&"
	case Eq:
		return "=="
	case Neq:
		return "!="
	case LParen:
		return "("
	case RParen:
		return ")"
	case Attr:
		return "attribute"
	case UUID:
		return "uuid"
	case EOF:
		return "EOF"
	}

	// Unreachable
	return ""
}

/*---------------------------------------------------------------------------*/

var (
	SupportedAttributes = []string{"user", "group", "tag", "event"}
	reAttrOrUUIDLike    = regexp.MustCompile(`^[a-z0-9\-:{}]+`)
)

func checkAttrOrUUIDLike(lexeme string) tokenKind {
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

	loc := reAttrOrUUIDLike.FindIndex(*b)

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

/*---------------------------------------------------------------------------*/

func createParseError(found tokenKind, expected ...tokenKind) error {
	return fmt.Errorf("expected %v, found %v", expected, found)
}

func consumeToken(ts *TokenStream, expected ...tokenKind) error {
	k := ts.Peek().Kind
	for _, e := range expected {
		if k == e {
			ts.Next()
			return nil
		}
	}
	return createParseError(k, expected...)
}

/*

top  : ε | expr
expr : term ( ( "||" | "&&" ) term)*
term : cmp | "(" expr ")"
cmp  : Attr ( "==" | "!=" ) UUID

*/

func CheckSyntax(ts *TokenStream) (err error) {
	if ts.HasNext() {
		err = checkSyntaxExpr(ts)
	}
	return
}

func checkSyntaxExpr(ts *TokenStream) error {
	if err := checkSyntaxTerm(ts); err != nil {
		return err
	}

	for ts.HasNext() {
		if err := consumeToken(ts, Or, And); err != nil {
			return err
		}
		if err := checkSyntaxTerm(ts); err != nil {
			return err
		}
	}

	return nil
}

func checkSyntaxTerm(ts *TokenStream) error {
	switch k := ts.Peek().Kind; k {
	case Attr:
		if err := checkSyntaxCmp(ts); err != nil {
			return err
		}

	case LParen:
		ts.Next()
		if err := checkSyntaxExpr(ts); err != nil {
			return err
		}
		if err := consumeToken(ts, RParen); err != nil {
			return err
		}

	default:
		return createParseError(k, Attr, LParen)
	}

	// Unreachable
	return nil
}

func checkSyntaxCmp(ts *TokenStream) error {
	if err := consumeToken(ts, Attr); err != nil {
		return err
	}
	if err := consumeToken(ts, Eq, Neq); err != nil {
		return err
	}
	if err := consumeToken(ts, UUID); err != nil {
		return err
	}
	return nil
}
