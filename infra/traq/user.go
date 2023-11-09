package traq

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/go-traq"
	"golang.org/x/oauth2"
)

func (repo *TraQRepository) GetUser(token *oauth2.Token, userID uuid.UUID) (*traq.User, error) {
	ctx := context.TODO()
	apiClient := MakeApiClient(token, ctx)
	userDetail, _, err := apiClient.UserApi.GetUser(ctx, userID.String()).Execute()
	if err != nil {
		return nil, err
	}
	user := new(traq.User)
	user.Id = userDetail.Id
	user.Name = userDetail.Name
	user.DisplayName = userDetail.DisplayName
	user.IconFileId = userDetail.IconFileId
	user.Bot = userDetail.Bot
	user.State = userDetail.State
	user.UpdatedAt = userDetail.UpdatedAt
	return user, err
}

func (repo *TraQRepository) GetUsers(token *oauth2.Token, includeSuspended bool) ([]traq.User, error) {
	ctx := context.TODO()
	apiClient := MakeApiClient(token, ctx)
	users, _, err := apiClient.UserApi.GetUsers(ctx).IncludeSuspended(includeSuspended).Execute()
	if err != nil {
		return nil, err
	}
	return users, err
}

func (repo *TraQRepository) GetUserMe(token *oauth2.Token) (*traq.User, error) {
	ctx := context.TODO()
	apiClient := MakeApiClient(token, ctx)

	userDetail, _, err := apiClient.MeApi.GetMe(ctx).Execute()
	if err != nil {
		return nil, err
	}
	user := new(traq.User)
	user.Id = userDetail.Id
	user.Name = userDetail.Name
	user.DisplayName = userDetail.DisplayName
	user.IconFileId = userDetail.IconFileId
	user.Bot = userDetail.Bot
	user.State = userDetail.State
	user.UpdatedAt = userDetail.UpdatedAt
	return user, err
}
