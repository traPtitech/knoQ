package parsing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*---------------------------------------------------------------------------*/

var lexCasesSuccess = []struct {
	in  string
	out *TokenStream
}{
	{"", NewTokenStream()},
	{"user group tag event", NewTokenStream(Token{Attr, "user"},
		Token{Attr, "group"}, Token{Attr, "tag"}, Token{Attr, "event"})},
	{"()&&||==!=", NewTokenStream(Token{LParen, ""}, Token{RParen, ""},
		Token{AndOp, ""}, Token{OrOp, ""}, Token{EqOp, ""}, Token{NeqOp, ""})},
	{"user==user&&tag==tag", NewTokenStream(Token{Attr, "user"}, Token{EqOp, ""},
		Token{Attr, "user"}, Token{AndOp, ""}, Token{Attr, "tag"},
		Token{EqOp, ""}, Token{Attr, "tag"})},
	{"123e4567-e89b-12d3-a456-426652340000",
		NewTokenStream(Token{UUID, "123e4567-e89b-12d3-a456-426652340000"})},
}

func TestLex_Success(t *testing.T) {
	t.Parallel()

	for _, c := range lexCasesSuccess {
		ts, err := Lex(c.in)
		assert.NoError(t, err)
		assert.Equal(t, c.out.tokens, ts.tokens)
	}
}

var lexCasesFailure = []struct {
	in string
}{
	{"#"},
	{"%"},
	{"useruser"},
	{"==="},
	{"123e4567--e89b-12d3-a456-426652340000"},
}

func TestLex_Failure(t *testing.T) {
	t.Parallel()

	for _, c := range lexCasesFailure {
		_, err := Lex(c.in)
		assert.Error(t, err)
	}
}

/*---------------------------------------------------------------------------*/

var parseCasesSuccess = []struct {
	in  string
	out Expr
}{
	{
		"",
		nil,
	},
	{
		"user==123e4567-e89b-12d3-a456-426652340000",
		&CmpExpr{"user", Eq, "123e4567-e89b-12d3-a456-426652340000"},
	},
	{
		"(((user==123e4567-e89b-12d3-a456-426652340000)))",
		&CmpExpr{"user", Eq, "123e4567-e89b-12d3-a456-426652340000"},
	},
	{
		"user==123e4567-e89b-12d3-a456-426652340000&&tag!=123e4567-e89b-12d3-a456-426652340000",
		&LogicOpExpr{
			And,
			&CmpExpr{"user", Eq, "123e4567-e89b-12d3-a456-426652340000"},
			&CmpExpr{"tag", Neq, "123e4567-e89b-12d3-a456-426652340000"},
		},
	},
	{
		"user==123e4567-e89b-12d3-a456-426652340000&&(tag!=123e4567-e89b-12d3-a456-426652340000||event==123e4567-e89b-12d3-a456-426652340000)",
		&LogicOpExpr{
			And,
			&CmpExpr{"user", Eq, "123e4567-e89b-12d3-a456-426652340000"},
			&LogicOpExpr{
				Or,
				&CmpExpr{"tag", Neq, "123e4567-e89b-12d3-a456-426652340000"},
				&CmpExpr{"event", Eq, "123e4567-e89b-12d3-a456-426652340000"},
			},
		},
	},
}

func TestParse_Success(t *testing.T) {
	t.Parallel()

	for _, c := range parseCasesSuccess {
		expr, err := Parse(c.in)
		assert.NoError(t, err)
		assert.Equal(t, c.out, expr)
	}
}

var parseCasesFailure = []struct {
	in string
}{
	{"user"},
	{"event=="},
	{"tag== || user==123e4567-e89b-12d3-a456-426652340000"},
	{"tag==123e4567-e89b-12d3-a456-426652340000||(user==123e4567-e89b-12d3-a456-426652340000))"},
}

func TestParse_Failure(t *testing.T) {
	t.Parallel()

	for _, c := range parseCasesFailure {
		_, err := Parse(c.in)
		assert.Error(t, err)
	}
}
