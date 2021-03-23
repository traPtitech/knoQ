package traq

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"

	traQ "github.com/traPtitech/traQ/router/v3"
)

func (repo *TraQRepository) GetGroup(token *oauth2.Token, groupID uuid.UUID) (*traQ.UserGroup, error) {
	URL := fmt.Sprintf("%s/groups/%s", repo.URL, groupID)
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	data, err := repo.doRequest(token, req)
	if err != nil {
		return nil, err
	}

	group := new(traQ.UserGroup)
	err = json.Unmarshal(data, &group)
	return group, err
}
