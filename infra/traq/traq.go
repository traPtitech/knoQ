package traq

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/url"

	"github.com/traPtitech/knoQ/utils/random"
	"golang.org/x/oauth2"
)

// TraQRepository is traq
type TraQRepository struct { //nolint:revive
	Config            *oauth2.Config
	URL               string
	ServerAccessToken string
}

func newPKCE() (pkceOptions []oauth2.AuthCodeOption, codeVerifier string) {
	codeVerifier = random.AlphaNumeric(43, true)
	result := sha256.Sum256([]byte(codeVerifier))
	enc := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_").WithPadding(base64.NoPadding)

	return []oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("code_challenge", enc.EncodeToString(result[:])),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		},
		codeVerifier
}

func (repo *TraQRepository) GetOAuthURL() (url, state, codeVerifier string) {
	pkceOptions, codeVerifier := newPKCE()
	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	state = random.AlphaNumeric(10, true)
	url = repo.Config.AuthCodeURL(state, pkceOptions...)
	return
}

func (repo *TraQRepository) GetOAuthToken(query, state, codeVerifier string) (*oauth2.Token, error) {
	ctx := context.TODO()
	values, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}
	if state != values.Get("state") {
		return nil, errors.New("state error")
	}
	code := values.Get("code")
	option := oauth2.SetAuthURLParam("code_verifier", codeVerifier)
	return repo.Config.Exchange(ctx, code, option)
}
