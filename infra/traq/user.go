package traq

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/go-traq"
	"github.com/traPtitech/knoQ/infra"

	"github.com/samber/lo"
)

func (repo *traqRepository) GetUser(userID uuid.UUID) (*infra.TraqUserResponse, error) {
	ctx := context.WithValue(context.TODO(), traq.ContextAccessToken, repo.ServerAccessToken)
	apiClient := traq.NewAPIClient(traq.NewConfiguration())
	// TODO: 一定期間キャッシュする
	u, resp, err := apiClient.UserApi.GetUser(ctx, userID.String()).Execute()
	if err != nil {
		return nil, err
	}
	err = handleStatusCode(resp.StatusCode)
	if err != nil {
		return nil, err
	}
	user := infra.TraqUserResponse{
		ID:          uuid.FromStringOrNil(u.Id),
		Name:        u.Name,
		DisplayName: u.DisplayName,
		IconURL:     "https://q.trap.jp/api/v3/public/icon/" + u.Name,
		Bot:         u.Bot,
		State:       u.State,
		UpdatedAt:   u.UpdatedAt,
	}
	return &user, err
}

func (repo *traqRepository) GetUsers(includeSuspended bool) ([]*infra.TraqUserResponse, error) {
	ctx := context.WithValue(context.TODO(), traq.ContextAccessToken, repo.ServerAccessToken)
	apiClient := traq.NewAPIClient(traq.NewConfiguration())
	// TODO: 一定期間キャッシュする
	us, resp, err := apiClient.UserApi.GetUsers(ctx).IncludeSuspended(includeSuspended).Execute()
	if err != nil {
		return nil, err
	}
	err = handleStatusCode(resp.StatusCode)
	if err != nil {
		return nil, err
	}

	users := lo.Map(us, func(u traq.User, _ int) *infra.TraqUserResponse {
		return &infra.TraqUserResponse{
			ID:          uuid.FromStringOrNil(u.Id),
			Name:        u.Name,
			DisplayName: u.DisplayName,
			IconURL:     "https://q.trap.jp/api/v3/public/icon/" + u.Name,
			Bot:         u.Bot,
			State:       u.State,
			UpdatedAt:   u.UpdatedAt,
		}
	})

	return users, err
}

func (repo *traqRepository) GetUserMe(accessToken string) (*infra.TraqUserResponse, error) {
	ctx := context.WithValue(context.TODO(), traq.ContextAccessToken, accessToken)
	apiClient := traq.NewAPIClient(traq.NewConfiguration())
	u, resp, err := apiClient.MeApi.GetMe(ctx).Execute()
	if err != nil {
		return nil, err
	}
	err = handleStatusCode(resp.StatusCode)
	if err != nil {
		return nil, err
	}

	user := infra.TraqUserResponse{
		ID:          uuid.FromStringOrNil(u.Id),
		Name:        u.Name,
		DisplayName: u.DisplayName,
		IconURL:     "https://q.trap.jp/api/v3/public/icon/" + u.Name,
		Bot:         u.Bot,
		State:       u.State,
		UpdatedAt:   u.UpdatedAt,
	}
	return &user, err
}
