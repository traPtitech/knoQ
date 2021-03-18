package google

import (
	"context"
	_ "embed"
	"errors"
	"net/url"

	"github.com/traPtitech/traQ/utils/random"
	"golang.org/x/oauth2"
)

type GoogleRepository struct {
	Config *oauth2.Config
}

//go:embed tmp/client.json
var ClientFile []byte

func (repo *GoogleRepository) GetOAuthURL() (url, state string) {
	state = random.SecureAlphaNumeric(10)
	url = repo.Config.AuthCodeURL(state)
	return
}

func (repo *GoogleRepository) GetOAuthToken(query, state string) (*oauth2.Token, error) {
	ctx := context.TODO()
	values, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}
	if state != values.Get("state") {
		return nil, errors.New("state error")
	}
	code := values.Get("code")
	return repo.Config.Exchange(ctx, code)
}
