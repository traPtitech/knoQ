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

func (s *service) SyncUsers(ctx context.Context, reqID uuid.UUID) error {
	if !s.IsPrivilege(ctx, reqID) {
		return domain.ErrForbidden
	}
	traQUsers, err := s.TraQRepo.GetUsers(true)
	if err != nil {
		return defaultErrorHandling(err)
	}

	args := make([]domain.SyncUserArgs, 0)
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
		args = append(args, a)
	}

	err = s.GormRepo.SyncUsers(args)
	return defaultErrorHandling(err)
}

func (s *service) GetOAuthURL(ctx context.Context) (url, state, codeVerifier string) {
	return s.TraQRepo.GetOAuthURL()
}

func (s *service) LoginUser(ctx context.Context, query, state, codeVerifier string) (*domain.User, error) {
	t, err := s.TraQRepo.GetOAuthToken(query, state, codeVerifier)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	traQUser, err := s.TraQRepo.GetUserMe(t)
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
	_, err = s.GormRepo.SaveUser(user)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	u, err := s.GetUser(ctx, user.UserID)
	return u, defaultErrorHandling(err)
}

func (s *service) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	userMeta, err := s.GormRepo.GetUser(userID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	if userMeta.Provider.Issuer == traQIssuerName {
		userBody, err := s.TraQRepo.GetUser(userID)
		if err != nil {
			return nil, defaultErrorHandling(err)
		}
		user, _ := s.mergeUser(userMeta, userBody)
		return user, nil
	}

	return nil, errors.New("not implemented")
}

func (s *service) GetUserMe(ctx context.Context, reqID uuid.UUID) (*domain.User, error) {
	return s.GetUser(ctx, reqID)
}

func (s *service) GetAllUsers(ctx context.Context, includeSuspend, includeBot bool) ([]*domain.User, error) {
	userMetas, err := s.GormRepo.GetAllUsers(!includeSuspend)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	// TODO fix
	traQUserBodys, err := s.TraQRepo.GetUsers(includeSuspend)
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
		user, err := s.mergeUser(userMeta, userBody)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

func (s *service) ReNewMyiCalSecret(ctx context.Context, reqID uuid.UUID) (secret string, err error) {
	secret = random.AlphaNumeric(16, true)
	err = s.GormRepo.UpdateiCalSecret(reqID, secret)
	return
}

func (s *service) GetMyiCalSecret(ctx context.Context, reqID uuid.UUID) (string, error) {
	user, err := s.GormRepo.GetUser(reqID)
	if err != nil {
		return "", defaultErrorHandling(err)
	}
	if user.State != 1 {
		return "", domain.ErrForbidden
	}
	// TODO: userid に対応する ical secret を返す関数を実装
	secret, err := s.GormRepo.GetICalSecret(reqID)
	if err != nil {
		return "", defaultErrorHandling(err)
	}
	if secret == "" {
		return "", domain.ErrNotFound
	}
	return secret, nil
}

func (s *service) IsPrivilege(ctx context.Context, reqID uuid.UUID) bool {
	user, err := s.GormRepo.GetUser(reqID)
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

func (s *service) mergeUser(userMeta *domain.User, userBody *traq.User) (*domain.User, error) {
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
		Icon:        s.TraQRepo.URL + "/public/icon/" + userBody.Name,
		Privileged:  userMeta.Privileged,
		State:       userMeta.State,
	}, nil
}

func (s *service) GrantPrivilege(ctx context.Context, userID uuid.UUID) error {
	user, err := s.GormRepo.GetUser(userID)
	if err != nil {
		return defaultErrorHandling(err)
	}
	if user.Privileged {
		return fmt.Errorf("%w: user has been already privileged", domain.ErrBadRequest)
	}
	err = s.GormRepo.GrantPrivilege(userID)
	return defaultErrorHandling(err)
}
