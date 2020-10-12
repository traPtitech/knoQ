package parsing

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*---------------------------------------------------------------------------*/

var lexCasesSuccess = []struct {
	in  string
	out TokenStream
}{
	{"", NewTokenStream()},
	{"user group tag event", NewTokenStream(Token{Attr, "user"},
		Token{Attr, "group"}, Token{Attr, "tag"}, Token{Attr, "event"})},
	{"()&&||==!=", NewTokenStream(Token{LParen, ""}, Token{RParen, ""},
		Token{And, ""}, Token{Or, ""}, Token{Eq, ""}, Token{Neq, ""})},
	{"user==user&&tag==tag", NewTokenStream(Token{Attr, "user"}, Token{Eq, ""},
		Token{Attr, "user"}, Token{And, ""}, Token{Attr, "tag"},
		Token{Eq, ""}, Token{Attr, "tag"})},
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

func tsOf(in ...tokenKind) TokenStream {
	var ts []Token
	for _, k := range in {
		ts = append(ts, Token{k, ""})
	}
	return NewTokenStream(ts...)
}

var chkSynCasesSuccess = []struct {
	in TokenStream
}{
	{tsOf()},
	{tsOf(Attr, Eq, UUID)},
	{tsOf(LParen, Attr, Eq, UUID, RParen)},
	{tsOf(Attr, Eq, UUID, And, Attr, Neq, UUID, Or, Attr, Eq, UUID)},
	{tsOf(LParen, LParen, Attr, Eq, UUID, Or, Attr, Eq, UUID, RParen, And,
		Attr, Neq, UUID, RParen, Or, Attr, Neq, UUID)},
}

func TestChkSyn_Success(t *testing.T) {
	t.Parallel()

	for _, c := range chkSynCasesSuccess {
		err := CheckSyntax(&c.in)
		assert.NoError(t, err)
	}
}

var chkSynCasesFailure = []struct {
	in TokenStream
}{
	{tsOf(Attr)},
	{tsOf(UUID)},
	{tsOf(And)},
	{tsOf(LParen, RParen)},
	{tsOf(LParen, Attr, Eq, UUID)},
}

func TestChkSyn_Failure(t *testing.T) {
	t.Parallel()

	for _, c := range chkSynCasesFailure {
		err := CheckSyntax(&c.in)
		assert.Error(t, err)
	}
}

func TestLexAndCheckSyntax(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    TokenStream
		wantErr bool
	}{
		{
			name: "illegal input",
			args: args{
				input: "aabbb===b",
			},
			want:    NewTokenStream(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LexAndCheckSyntax(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("LexAndCheckSyntax() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LexAndCheckSyntax() = %v, want %v", got, tt.want)
			}
		})
	}
}