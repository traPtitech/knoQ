package traq

import (
	"context"
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

	// ここから go-traq 書き換え

	// traqconf := traq.NewConfiguration()
	// conf := oauth2.Config{

	// }
	// traqconf.HTTPClient = conf.Client(context.TODO(),token)
	// apiClient := traq.NewAPIClient(traqconf)
	// userDtail,_,err := apiClient.UserApi.GetUser(context.Background(),userID.String()).Execute()
	// if err!=nil{
	// 	return nil,err
	// }
	// user := new(traq.User)
	// return user, err

	// ここまで go-traq 書き換え
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
