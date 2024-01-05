package traq

import (
	"context"

	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"

	"github.com/traPtitech/go-traq"
)

func (repo *TraQRepository) GetGroup(token *oauth2.Token, groupID uuid.UUID) (*traq.UserGroup, error) {
	ctx := context.TODO()
	apiClient := NewAPIClient(ctx, token)
	group, resp, err := apiClient.GroupApi.GetUserGroup(ctx, groupID.String()).Execute()
	if err != nil {
		return nil, err
	}
	err = handleStatusCode(resp.StatusCode)
	if err != nil {
		return nil, err
	}
	return group, err
}

func (repo *TraQRepository) GetAllGroups(token *oauth2.Token) ([]*traq.UserGroup, error) {
	ctx := context.TODO()
	apiClient := NewAPIClient(ctx, token)
	groups, resp, err := apiClient.GroupApi.GetUserGroups(ctx).Execute()
	if err != nil {
		return nil, err
	}
	err = handleStatusCode(resp.StatusCode)
	if err != nil {
		return nil, err
	}
	ret := make([]*traq.UserGroup, len(groups))
	for i, _g := range groups {
		g := _g
		ret[i] = &g
	}
	return ret, err
}

func (repo *TraQRepository) GetUserBelongingGroupIDs(token *oauth2.Token, userID uuid.UUID) ([]uuid.UUID, error) {
	ctx := context.TODO()
	apiClient := NewAPIClient(ctx, token)
	user, resp, err := apiClient.UserApi.GetUser(ctx, userID.String()).Execute()
	if err != nil {
		return nil, err
	}
	err = handleStatusCode(resp.StatusCode)
	if err != nil {
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
