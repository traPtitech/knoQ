package service

import (
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/go-traq"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
	"github.com/traPtitech/knoQ/utils/random"
)

const traQIssuerName = "traQ"

func (repo *service) SyncUsers(info *domain.ConInfo) error {
	if !repo.IsPrivilege(info) {
		return domain.ErrForbidden
	}
	traQUsers, err := repo.TraQRepo.GetUsers(true)
	if err != nil {
		return defaultErrorHandling(err)
	}

	users := make([]*db.User, 0)
	for _, u := range traQUsers {
		if u.Bot {
			continue
		}

		uid := uuid.Must(uuid.FromString(u.GetId()))
		user := &db.User{
			ID:    uid,
			State: int(u.State),
			Provider: db.Provider{
				UserID:  uid,
				Issuer:  traQIssuerName,
				Subject: u.GetId(),
			},
		}
		users = append(users, user)
	}

	err = repo.GormRepo.SyncUsers(users)
	return defaultErrorHandling(err)
}

func (repo *service) GetOAuthURL() (url, state, codeVerifier string) {
	return repo.TraQRepo.GetOAuthURL()
}

func (repo *service) LoginUser(query, state, codeVerifier string) (*domain.User, error) {
	t, err := repo.TraQRepo.GetOAuthToken(query, state, codeVerifier)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	traQUser, err := repo.TraQRepo.GetUserMe(t)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	uid := uuid.Must(uuid.FromString(traQUser.GetId()))
	user := db.User{
		ID:    uid,
		State: 1,
		Token: db.Token{
			UserID: uid,
			Oauth2Token: &db.Oauth2Token{
				AccessToken:  t.AccessToken,
				TokenType:    t.TokenType,
				RefreshToken: t.RefreshToken,
				Expiry:       t.Expiry,
			},
		},
		Provider: db.Provider{
			UserID:  uid,
			Issuer:  traQIssuerName,
			Subject: traQUser.GetId(),
		},
	}
	_, err = repo.GormRepo.SaveUser(user)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	u, err := repo.GetUser(user.ID, &domain.ConInfo{
		ReqUserID: user.ID,
	})
	return u, defaultErrorHandling(err)
}

func (repo *service) GetUser(userID uuid.UUID, _ *domain.ConInfo) (*domain.User, error) {
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

func (repo *service) GetUserMe(info *domain.ConInfo) (*domain.User, error) {
	return repo.GetUser(info.ReqUserID, info)
}

func (repo *service) GetAllUsers(includeSuspend, includeBot bool, _ *domain.ConInfo) ([]*domain.User, error) {
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

func (repo *service) ReNewMyiCalSecret(info *domain.ConInfo) (secret string, err error) {
	secret = random.AlphaNumeric(16, true)
	err = repo.GormRepo.UpdateiCalSecret(info.ReqUserID, secret)
	return
}

func (repo *service) GetMyiCalSecret(info *domain.ConInfo) (string, error) {
	user, err := repo.GormRepo.GetUser(info.ReqUserID)
	if err != nil {
		return "", defaultErrorHandling(err)
	}
	if user.State != 1 {
		return "", domain.ErrForbidden
	}
	if user.IcalSecret == "" {
		return "", domain.ErrNotFound
	}
	return user.IcalSecret, nil
}

func (repo *service) IsPrivilege(info *domain.ConInfo) bool {
	user, err := repo.GormRepo.GetUser(info.ReqUserID)
	if err != nil {
		return false
	}
	return user.Privilege
}

func traQUserMap(users []traq.User) map[uuid.UUID]*traq.User {
	userMap := make(map[uuid.UUID]*traq.User)
	for _, u := range users {
		user := u
		userMap[uuid.Must(uuid.FromString(user.GetId()))] = &user
	}
	return userMap
}

func (repo *service) mergeUser(userMeta *db.User, userBody *traq.User) (*domain.User, error) {
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
		Privileged:  userMeta.Privilege,
		State:       userMeta.State,
	}, nil
}

func (repo *service) GrantPrivilege(userID uuid.UUID) error {
	user, err := repo.GormRepo.GetUser(userID)
	if err != nil {
		return defaultErrorHandling(err)
	}
	if user.Privilege {
		return fmt.Errorf("%w: user has been already privileged", domain.ErrBadRequest)
	}
	err = repo.GormRepo.GrantPrivilege(userID)
	return defaultErrorHandling(err)
}
