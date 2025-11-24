package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/go-traq"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/utils/random"
)

const traQIssuerName = "traQ"

func (repo *service) SyncUsers(ctx context.Context) error {
	if !repo.IsPrivilege(ctx) {
		return domain.ErrForbidden
	}
	traQUsers, err := repo.TraQRepo.GetUsers(true)
	if err != nil {
		return defaultErrorHandling(err)
	}

	argss := make([]domain.SyncUserArgs, 0)
	for _, u := range traQUsers {
		if u.Bot {
			continue
		}

		uid := uuid.Must(uuid.FromString(u.GetId()))
		a := domain.SyncUserArgs{
			UserID: uid,
			State:  int(u.State),
			ProviderArgs: domain.ProviderArgs{
				Issuer:  traQIssuerName,
				Subject: u.GetId(),
			},
		}
		argss = append(argss, a)
	}

	err = repo.GormRepo.SyncUsers(argss)
	return defaultErrorHandling(err)
}

func (repo *service) GetOAuthURL(ctx context.Context) (url, state, codeVerifier string) {
	return repo.TraQRepo.GetOAuthURL()
}

func (repo *service) LoginUser(ctx context.Context, query, state, codeVerifier string) (*domain.User, error) {
	t, err := repo.TraQRepo.GetOAuthToken(query, state, codeVerifier)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	traQUser, err := repo.TraQRepo.GetUserMe(t)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	uid := uuid.Must(uuid.FromString(traQUser.GetId()))

	user := domain.SaveUserArgs{
		UserID: uid,
		State:  1,
		TokenArgs: domain.TokenArgs{
			AccessToken:  t.AccessToken,
			TokenType:    t.TokenType,
			RefreshToken: t.RefreshToken,
			Expiry:       t.Expiry,
		},
		ProviderArgs: domain.ProviderArgs{
			Issuer:  traQIssuerName,
			Subject: traQUser.GetId(),
		},
	}
	_, err = repo.GormRepo.SaveUser(user)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	u, err := repo.GetUser(ctx, user.UserID)
	return u, defaultErrorHandling(err)
}

func (repo *service) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	userMeta, err := repo.GormRepo.GetUser(userID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	if userMeta.Provider.Issuer == traQIssuerName {
		userBody, err := repo.TraQRepo.GetUser(userID)
		if err != nil {
			return nil, defaultErrorHandling(err)
		}
		user, _ := repo.mergeUser(userMeta, userBody)
		return user, nil
	}
	// userBody, err := repo.gormRepo.GetUserBody(userID)

	return nil, errors.New("not implemented")
}

func (repo *service) GetUserMe(ctx context.Context) (*domain.User, error) {
	reqID, _ := domain.GetUserID(ctx)
	return repo.GetUser(ctx, reqID)
}

func (repo *service) GetAllUsers(ctx context.Context, includeSuspend, includeBot bool) ([]*domain.User, error) {
	userMetas, err := repo.GormRepo.GetAllUsers(!includeSuspend)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	// TODO fix
	traQUserBodys, err := repo.TraQRepo.GetUsers(includeSuspend)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	traQUserBodsMap := traQUserMap(traQUserBodys)
	users := make([]*domain.User, 0, len(userMetas))
	for _, userMeta := range userMetas {
		userBody, ok := traQUserBodsMap[userMeta.ID]
		if !ok {
			continue
		}
		if !includeBot && userBody.Bot {
			continue
		}
		user, err := repo.mergeUser(userMeta, userBody)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

func (repo *service) ReNewMyiCalSecret(ctx context.Context) (secret string, err error) {
	secret = random.AlphaNumeric(16, true)
	reqID, _ := domain.GetUserID(ctx)
	err = repo.GormRepo.UpdateiCalSecret(reqID, secret)
	return
}

func (repo *service) GetMyiCalSecret(ctx context.Context) (string, error) {
	reqID, _ := domain.GetUserID(ctx)
	user, err := repo.GormRepo.GetUser(reqID)
	if err != nil {
		return "", defaultErrorHandling(err)
	}
	if user.State != 1 {
		return "", domain.ErrForbidden
	}
	// TODO: userid に対応する ical secret を返す関数を実装
	secret, err := repo.GormRepo.GetICalSecret(reqID)
	if err != nil {
		return "", defaultErrorHandling(err)
	}
	if secret == "" {
		return "", domain.ErrNotFound
	}
	return secret, nil
}

func (repo *service) IsPrivilege(ctx context.Context) bool {
	reqID, _ := domain.GetUserID(ctx)

	user, err := repo.GormRepo.GetUser(reqID)
	if err != nil {
		return false
	}
	return user.Privileged
}

func traQUserMap(users []traq.User) map[uuid.UUID]*traq.User {
	userMap := make(map[uuid.UUID]*traq.User)
	for _, u := range users {
		user := u
		userMap[uuid.Must(uuid.FromString(user.GetId()))] = &user
	}
	return userMap
}

func (repo *service) mergeUser(userMeta *domain.User, userBody *traq.User) (*domain.User, error) {
	if userMeta.ID != uuid.Must(uuid.FromString(userBody.GetId())) {
		return nil, errors.New("id does not match")
	}
	if userMeta.Provider.Issuer != traQIssuerName {
		return nil, errors.New("different provider")
	}
	return &domain.User{
		ID:          userMeta.ID,
		Name:        userBody.Name,
		DisplayName: userBody.DisplayName,
		Icon:        repo.TraQRepo.URL + "/public/icon/" + userBody.Name,
		Privileged:  userMeta.Privileged,
		State:       userMeta.State,
	}, nil
}

func (repo *service) GrantPrivilege(ctx context.Context, userID uuid.UUID) error {
	user, err := repo.GormRepo.GetUser(userID)
	if err != nil {
		return defaultErrorHandling(err)
	}
	if user.Privileged {
		return fmt.Errorf("%w: user has been already privileged", domain.ErrBadRequest)
	}
	err = repo.GormRepo.GrantPrivilege(userID)
	return defaultErrorHandling(err)
}
