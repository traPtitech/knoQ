package traq

import (
	"context"
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/oauth2"

	traQrandom "github.com/traPtitech/traQ/utils/random"
)

// TraQRepository is traq
type TraQRepository struct {
	Config *oauth2.Config
}

var TraQDefaultConfig = &oauth2.Config{
	ClientID:     "something",
	ClientSecret: "any",
	RedirectURL:  "foo",
	Scopes:       []string{"read", "write", "manage_bot"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://q.trap.jp/api/v3/oauth2/authorize",
		TokenURL: "https://q.trap.jp/api/v3/oauth2/token",
	},
}

func newPKCE() (pkceOptions []oauth2.AuthCodeOption, codeVerifier string) {
	codeVerifier = traQrandom.SecureAlphaNumeric(43)
	result := sha256.Sum256([]byte(codeVerifier))
	enc := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_").WithPadding(base64.NoPadding)

	return []oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("code_challenge", enc.EncodeToString(result[:])),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		},
		codeVerifier
}

func (repo *TraQRepository) GetOAuthURL() (url, codeVerifier string) {
	pkceOptions, codeVerifier := newPKCE()
	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	state := traQrandom.SecureAlphaNumeric(10)
	url = repo.Config.AuthCodeURL(state, pkceOptions...)
	return
}

func (repo *TraQRepository) GetOAuthToken(code, codeVerifier string) (*oauth2.Token, error) {
	ctx := context.TODO()
	option := oauth2.SetAuthURLParam("code_verifier", codeVerifier)
	return repo.Config.Exchange(ctx, code, option)
}
