package parsing

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain/filter"
)

/*---------------------------------------------------------------------------*/

func Parse(input string) (filter.Expr, error) {
	ts, err := Lex(input)
	if err != nil {
		return nil, err
	}

	expr, err := ParseTop(ts)
	if err != nil {
		return nil, err
	}

	return expr, nil
}

/*---------------------------------------------------------------------------*/

// TokenStream is peekable and restorable stream of Tokens
type TokenStream struct {
	tokens []Token
	pos    int
}

// NewTokenStream creates new TokenStream of given tokens
func NewTokenStream(tokens ...Token) *TokenStream {
	tokens = append(tokens, Token{EOF, ""})
	return &TokenStream{tokens, 0}
}

// HasNext checks if ts has next token
func (ts *TokenStream) HasNext() bool {
	return ts.pos <= len(ts.tokens)-1
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
	OrOp
	AndOp
	EqOp
	Like
	NeqOp
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
	case OrOp:
		s = "||"
	case AndOp:
		s = "&&"
	case EqOp:
		s = "=="
	case Like:
		s = "Like"
	case NeqOp:
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
func Lex(input string) (*TokenStream, error) {
	bytes := []byte(input)
	var tokens []Token

	for len(bytes) > 0 {
		token, err := advanceToken(&bytes)
		if err != nil {
			return nil, err
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
		token = Token{OrOp, ""}
		*b = (*b)[2:]

	case bytes.HasPrefix(*b, []byte("&&")):
		token = Token{AndOp, ""}
		*b = (*b)[2:]

	case bytes.HasPrefix(*b, []byte("==")):
		token = Token{EqOp, ""}
		*b = (*b)[2:]

	case bytes.HasPrefix(*b, []byte("!=")):
		token = Token{NeqOp, ""}
		*b = (*b)[2:]

	case bytes.HasPrefix(*b, []byte("Like")):
		token = Token{Like, ""}
		*b = (*b)[4:]

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

/*---------------------------------------------------------------------------*/

var MapSupportedAttributes = map[string]filter.Attr{
	"user":  filter.AttrUser,
	"group": filter.AttrGroup,
	"tag":   filter.AttrTag,
	"event": filter.AttrEvent,
}

func createParseError(found tokenKind, expected ...tokenKind) error {
	return fmt.Errorf("expected %v, found %v", expected, found)
}

func consumeToken(ts *TokenStream, expected ...tokenKind) error {
	k := ts.Next().Kind
	for _, e := range expected {
		if k == e {
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

func ParseTop(ts *TokenStream) (filter.Expr, error) {
	var expr filter.Expr
	var err error

	if ts.Peek().Kind != EOF {
		if expr, err = ParseExpr(ts); err != nil {
			return nil, err
		}
	}

	if err = consumeToken(ts, EOF); err != nil {
		return nil, err
	}

	return expr, nil
}

func ParseExpr(ts *TokenStream) (filter.Expr, error) {
	var expr filter.Expr
	var err error

	if expr, err = ParseTerm(ts); err != nil {
		return nil, err
	}

Loop:
	for {
		switch k := ts.Peek().Kind; k {
		case AndOp, OrOp:
			ts.Next()
			lhs := expr
			op := map[tokenKind]filter.LogicOp{
				AndOp: filter.And,
				OrOp:  filter.Or,
			}[k]
			rhs, err := ParseTerm(ts)
			if err != nil {
				return nil, err
			}
			expr = &filter.LogicOpExpr{op, lhs, rhs}

		default:
			break Loop
		}
	}

	return expr, nil
}

func ParseTerm(ts *TokenStream) (filter.Expr, error) {
	var expr filter.Expr
	var err error

	switch k := ts.Peek().Kind; k {
	case Attr:
		if expr, err = ParseCmp(ts); err != nil {
			return nil, err
		}

	case LParen:
		ts.Next()
		if expr, err = ParseExpr(ts); err != nil {
			return nil, err
		}
		if err = consumeToken(ts, RParen); err != nil {
			return nil, err
		}

	default:
		return nil, createParseError(k, Attr, LParen)
	}

	return expr, nil
}

func ParseCmp(ts *TokenStream) (filter.Expr, error) {
	var attr string
	var rel filter.Relation
	var uid uuid.UUID

	tok := ts.Next()
	if tok.Kind != Attr {
		return nil, createParseError(tok.Kind, Attr)
	}
	attr = tok.Value

	tok = ts.Next()
	if tok.Kind != EqOp && tok.Kind != NeqOp && tok.Kind != Like {
		return nil, createParseError(tok.Kind, EqOp, NeqOp)
	}
	rel = map[tokenKind]filter.Relation{
		EqOp:  filter.Eq,
		NeqOp: filter.Neq,
		Like:  filter.Like,
	}[tok.Kind]

	tok = ts.Next()
	if tok.Kind != UUID {
		return nil, createParseError(tok.Kind, UUID)
	}
	uid = uuid.Must(uuid.FromString(tok.Value))

	return &filter.CmpExpr{MapSupportedAttributes[attr], rel, uid}, nil
}
