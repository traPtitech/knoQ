package traq

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/go-traq"
	"golang.org/x/oauth2"
)

func (repo *TraQRepository) GetUser(token *oauth2.Token, userID uuid.UUID) (*traq.User, error) {
	apiClient := MakeApiClient(token, context.Background())
	userDetail, _, err := apiClient.UserApi.GetUser(context.Background(), userID.String()).Execute()
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

func (repo *TraQRepository) GetUsers(token *oauth2.Token, includeSuspended bool) ([]*traq.User, error) {
	apiClient := MakeApiClient(token, context.Background())
	users, _, err := apiClient.UserApi.GetUsers(context.Background()).IncludeSuspended(includeSuspended).Execute()
	if err != nil {
		return nil, err
	}
	// return users,err
	res_users := convertUsersToUsers(users)
	return res_users, err
}

func (repo *TraQRepository) GetUserMe(token *oauth2.Token) (*traq.User, error) {
	apiClient := MakeApiClient(token, context.Background())

	userDetail, _, err := apiClient.MeApi.GetMe(context.Background()).Execute()
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

func convertUsersToUsers(users []traq.User) []*traq.User {
	new_users := make([]*traq.User, len(users))
	for i, _user := range users {
		user := _user
		new_users[i] = &user
	}
	return new_users
}
