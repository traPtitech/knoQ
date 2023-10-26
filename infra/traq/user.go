package traq

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/go-traq"
	"golang.org/x/oauth2"
)

func (repo *TraQRepository) GetUser(token *oauth2.Token, userID uuid.UUID) (*traq.User, error) {
	apiClient := MakeApiClient(token)
	userDetail, _, err := apiClient.UserApi.GetUser(context.Background(), userID.String()).Execute()
	if err != nil {
		return nil, err
	}
	user := convertUserdetailToUser(userDetail)
	return user, err
}

func (repo *TraQRepository) GetUsers(token *oauth2.Token, includeSuspended bool) ([]*traq.User, error) {
	apiClient := MakeApiClient(token)
	users, _, err := apiClient.UserApi.GetUsers(context.Background()).IncludeSuspended(includeSuspended).Execute()
	if err != nil {
		return nil, err
	}
	res_users := convertUsersToUsers(users)
	return res_users, err
}

func (repo *TraQRepository) GetUserMe(token *oauth2.Token) (*traq.User, error) {
	apiClient := MakeApiClient(token)

	data, _, err := apiClient.MeApi.GetMe(context.Background()).Execute()
	if err != nil {
		return nil, err
	}
	user := convertMyUserdetailToUser(data)
	return user, err
}

func convertMyUserdetailToUser(userdetail *traq.MyUserDetail) *traq.User {
	user := new(traq.User)
	user.Id = userdetail.Id
	user.Name = userdetail.Name
	user.DisplayName = userdetail.DisplayName
	user.IconFileId = userdetail.IconFileId
	user.Bot = userdetail.Bot
	user.State = userdetail.State
	user.UpdatedAt = userdetail.UpdatedAt
	return user
}

func convertUserdetailToUser(userdetail *traq.UserDetail) *traq.User {
	user := new(traq.User)
	user.Id = userdetail.Id
	user.Name = userdetail.Name
	user.DisplayName = userdetail.DisplayName
	user.IconFileId = userdetail.IconFileId
	user.Bot = userdetail.Bot
	user.State = userdetail.State
	user.UpdatedAt = userdetail.UpdatedAt
	return user
}
func convertUsersToUsers(users []traq.User) []*traq.User {
	new_users := make([]*traq.User, len(users))
	for i, _user := range users {
		user := _user
		new_users[i] = &user
	}
	return new_users
}
