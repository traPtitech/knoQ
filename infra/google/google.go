package google

import (
	"context"
	"errors"
	"net/url"

	"github.com/traPtitech/knoQ/utils/random"
	"golang.org/x/oauth2"
)

type Repository struct {
	Config *oauth2.Config
}

// "embed" を利用したい場合以下のコメントを解除する
// import "embed"
// // go:embed tmp/client.json
// var ClientFile []byte

func (repo *Repository) GetOAuthURL() (url, state string) {
	state = random.AlphaNumeric(10, true)
	url = repo.Config.AuthCodeURL(state)
	return
}

func (repo *Repository) GetOAuthToken(query, state string) (*oauth2.Token, error) {
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
