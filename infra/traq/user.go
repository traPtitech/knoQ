package traq

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"

	traQ "github.com/traPtitech/traQ/router/v3"
)

func (repo *TraQRepository) doRequest(token *oauth2.Token, req *http.Request) ([]byte, error) {
	client := repo.Config.Client(context.TODO(), token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func (repo *TraQRepository) GetUser(token *oauth2.Token, userID uuid.UUID) (*traQ.User, error) {
	URL := fmt.Sprintf("%s/users/%s", repo.URL, userID)
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	data, err := repo.doRequest(token, req)
	if err != nil {
		return nil, err
	}

	user := new(traQ.User)
	err = json.Unmarshal(data, &user)
	return user, err
}

func (repo *TraQRepository) GetUsers(token *oauth2.Token, includeSuspended bool) ([]*traQ.User, error) {
	URL, err := url.Parse(fmt.Sprintf("%s/users", repo.URL))
	if err != nil {
		return nil, err
	}
	URL.Query().Set("include-suspended", strconv.FormatBool(includeSuspended))
	req, err := http.NewRequest(http.MethodGet, URL.String(), nil)
	if err != nil {
		return nil, err
	}
	data, err := repo.doRequest(token, req)
	if err != nil {
		return nil, err
	}

	users := make([]*traQ.User, 0)
	err = json.Unmarshal(data, &users)
	return users, err
}

func (repo *TraQRepository) GetUserMe(token *oauth2.Token) (*traQ.User, error) {
	URL := fmt.Sprintf("%s/users/me", repo.URL)
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	data, err := repo.doRequest(token, req)
	if err != nil {
		return nil, err
	}

	user := new(traQ.User)
	err = json.Unmarshal(data, &user)
	return user, err
}
