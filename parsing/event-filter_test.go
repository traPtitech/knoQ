package parsing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var casesSuccess = []struct {
	in  string
	out TokenStream
}{
	{"", NewTokenStream()},
	{"user", NewTokenStream(Token{Attr, "user"})},
	{"group", NewTokenStream(Token{Attr, "group"})},
	{"tag", NewTokenStream(Token{Attr, "tag"})},
	{"event", NewTokenStream(Token{Attr, "event"})},
	{"( ) && || == !=", NewTokenStream(Token{LParen, ""}, Token{RParen, ""},
		Token{And, ""}, Token{Or, ""}, Token{Eq, ""}, Token{Neq, ""})},
	{"user==user&&tag==tag", NewTokenStream(Token{Attr, "user"}, Token{Eq, ""},
		Token{Attr, "user"}, Token{And, ""}, Token{Attr, "tag"},
		Token{Eq, ""}, Token{Attr, "tag"})},
}

func TestLex_Success(t *testing.T) {
	t.Parallel()

	for _, c := range casesSuccess {
		ts, err := Lex(c.in)
		assert.Nil(t, err)
		assert.Equal(t, c.out.tokens, ts.tokens)
	}
}

var casesFailure = []struct {
	in string
}{
	{"#"},
}

func TestLex_Failure(t *testing.T) {
	t.Parallel()

	for _, c := range casesFailure {
		_, err := Lex(c.in)
		assert.NotNil(t, err)
	}
}
