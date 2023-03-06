package traq

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/go-traq"
	"golang.org/x/oauth2"
)

func (repo *TraQRepository) GetUser(token *oauth2.Token, userID uuid.UUID) (*traq.User, error) {
	URL := fmt.Sprintf("%s/users/%s", repo.URL, userID)
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	data, err := repo.doRequest(token, req)
	if err != nil {
		return nil, err
	}

	user := new(traq.User)
	err = json.Unmarshal(data, &user)
	return user, err
}

func (repo *TraQRepository) GetUsers(token *oauth2.Token, includeSuspended bool) ([]*traq.User, error) {
	URL, err := url.Parse(fmt.Sprintf("%s/users", repo.URL))
	if err != nil {
		return nil, err
	}
	q := URL.Query()
	q.Set("include-suspended", strconv.FormatBool(includeSuspended))
	URL.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodGet, URL.String(), nil)
	if err != nil {
		return nil, err
	}
	data, err := repo.doRequest(token, req)
	if err != nil {
		return nil, err
	}

	users := make([]*traq.User, 0)
	err = json.Unmarshal(data, &users)
	return users, err
}

func (repo *TraQRepository) GetUserMe(token *oauth2.Token) (*traq.User, error) {
	URL := fmt.Sprintf("%s/users/me", repo.URL)
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	data, err := repo.doRequest(token, req)
	if err != nil {
		return nil, err
	}

	user := new(traq.User)
	err = json.Unmarshal(data, &user)
	return user, err
}
