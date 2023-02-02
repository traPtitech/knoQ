package traq

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"

	"github.com/traPtitech/go-traq"
)

func (repo *TraQRepository) GetGroup(token *oauth2.Token, groupID uuid.UUID) (*traq.UserGroup, error) {
	URL := fmt.Sprintf("%s/groups/%s", repo.URL, groupID)
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	data, err := repo.doRequest(token, req)
	if err != nil {
		return nil, err
	}

	group := new(traq.UserGroup)
	err = json.Unmarshal(data, &group)
	return group, err
}

func (repo *TraQRepository) GetAllGroups(token *oauth2.Token) ([]*traq.UserGroup, error) {
	URL := fmt.Sprintf("%s/groups", repo.URL)
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	data, err := repo.doRequest(token, req)
	if err != nil {
		return nil, err
	}

	groups := make([]*traq.UserGroup, 0)
	err = json.Unmarshal(data, &groups)
	return groups, err
}

func (repo *TraQRepository) GetUserBelongingGroupIDs(token *oauth2.Token, userID uuid.UUID) ([]uuid.UUID, error) {
	URL := fmt.Sprintf("%s/users/%s", repo.URL, userID)
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	data, err := repo.doRequest(token, req)
	if err != nil {
		return nil, err
	}
	user := new(traq.UserDetail)
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	groups := make([]uuid.UUID, 0, len(user.Groups))
	for _, gid := range user.Groups {
		groupUUID, err := uuid.FromString(gid)
		if err != nil {
			return nil, err
		}
		groups = append(groups, groupUUID)
	}

	return groups, err
}
