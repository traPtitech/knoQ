package parsing

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"

	"github.com/gofrs/uuid"
)

/*---------------------------------------------------------------------------*/

// LexAndCheckSyntax tokenizes input and check syntax
// if input is illegal, this function returns error
func LexAndCheckSyntax(input string) (TokenStream, error) {
	ts, err := Lex(input)
	if err != nil {
		return NewTokenStream(), err
	}
	if err = CheckSyntax(&ts); err != nil {
		return NewTokenStream(), err
	}
	ts.Restore()
	return ts, nil
}

/*---------------------------------------------------------------------------*/

// TokenStream is peekable and restorable stream of Tokens
type TokenStream struct {
	tokens []Token
	pos    int
}

// NewTokenStream creates new TokenStream of given tokens
func NewTokenStream(tokens ...Token) TokenStream {
	tokens = append(tokens, Token{EOF, ""})
	return TokenStream{tokens, 0}
}

// HasNext checks if ts has next token
func (ts *TokenStream) HasNext() bool {
	return ts.tokens[ts.pos].Kind != EOF
}

// Next returns next token and proceeds
func (ts *TokenStream) Next() Token {
	t := ts.tokens[ts.pos]
	ts.pos++
	return t
}

// Peek returns next token without proceeding
func (ts *TokenStream) Peek() Token {
	return ts.tokens[ts.pos]
}

// Restore restores all the tokens consumed so far
func (ts *TokenStream) Restore() {
	ts.pos = 0
}

// Token has two fields: Kind and Value
// Value is used for holding attributes or UUID
// UUID Value has a canonical RFC-4122 string representation:
//     xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
type Token struct {
	Kind  tokenKind
	Value string
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

func (k tokenKind) String() (s string) {
	switch k {
	case Unknown:
		s = "unknown"
	case Or:
		s = "||"
	case And:
		s = "&&"
	case Eq:
		s = "=="
	case Neq:
		s = "!="
	case LParen:
		s = "("
	case RParen:
		s = ")"
	case Attr:
		s = "attribute"
	case UUID:
		s = "uuid"
	case EOF:
		s = "EOF"
	}
	return
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

// Lex tokenizes input and returns TokenStream and error
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

Syntax:
	top  : ε | expr
	expr : term ( ( "||" | "&&" ) term)*
	term : cmp | "(" expr ")"
	cmp  : Attr ( "==" | "!=" ) UUID

*/

// CheckSyntax checks if given TokenStream satisfies the syntax
// shown above
func CheckSyntax(ts *TokenStream) error {
	if ts.HasNext() {
		if err := checkSyntaxExpr(ts); err != nil {
			return err
		}
	}
	if ts.HasNext() {
		return createParseError(ts.Peek().Kind, EOF)
	}
	return nil
}

func checkSyntaxExpr(ts *TokenStream) error {
	if err := checkSyntaxTerm(ts); err != nil {
		return err
	}

	for ts.HasNext() {
		switch k := ts.Peek().Kind; k {
		case Or, And:
			ts.Next()
			if err := checkSyntaxTerm(ts); err != nil {
				return err
			}

		case RParen:
			return nil

		default:
			return createParseError(k, Or, And, RParen)
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
