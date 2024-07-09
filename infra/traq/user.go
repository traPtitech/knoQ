package traq

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/go-traq"
	"golang.org/x/oauth2"
)

func (repo *TraQRepository) GetUser(userID uuid.UUID) (*traq.User, error) {
	ctx := context.WithValue(context.TODO(), traq.ContextAccessToken, repo.ServerAccessToken)
	apiClient := traq.NewAPIClient(traqAPIConfig)
	// TODO: 一定期間キャッシュする
	userDetail, resp, err := apiClient.UserApi.GetUser(ctx, userID.String()).Execute()
	if err != nil {
		return nil, err
	}
	err = handleStatusCode(resp.StatusCode)
	if err != nil {
		return nil, err
	}
	user := traq.User{
		Id:          userDetail.Id,
		Name:        userDetail.Name,
		DisplayName: userDetail.DisplayName,
		IconFileId:  userDetail.IconFileId,
		Bot:         userDetail.Bot,
		State:       userDetail.State,
		UpdatedAt:   userDetail.UpdatedAt,
	}
	return &user, err
}

func (repo *TraQRepository) GetUsers(includeSuspended bool) ([]traq.User, error) {
	ctx := context.WithValue(context.TODO(), traq.ContextAccessToken, repo.ServerAccessToken)
	apiClient := traq.NewAPIClient(traqAPIConfig)
	// TODO: 一定期間キャッシュする
	users, resp, err := apiClient.UserApi.GetUsers(ctx).IncludeSuspended(includeSuspended).Execute()
	if err != nil {
		return nil, err
	}
	err = handleStatusCode(resp.StatusCode)
	if err != nil {
		return nil, err
	}
	return users, err
}

func (repo *TraQRepository) GetUserMe(token *oauth2.Token) (*traq.User, error) {
	ctx := context.TODO()
	apiClient := NewOauth2APIClient(ctx, token)
	userDetail, resp, err := apiClient.MeApi.GetMe(ctx).Execute()
	if err != nil {
		return nil, err
	}
	err = handleStatusCode(resp.StatusCode)
	if err != nil {
		return nil, err
	}
	user := traq.User{
		Id:          userDetail.Id,
		Name:        userDetail.Name,
		DisplayName: userDetail.DisplayName,
		IconFileId:  userDetail.IconFileId,
		Bot:         userDetail.Bot,
		State:       userDetail.State,
		UpdatedAt:   userDetail.UpdatedAt,
	}

	return &user, err
}
