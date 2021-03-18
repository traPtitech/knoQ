package traq

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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
