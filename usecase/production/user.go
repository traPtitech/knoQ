package production

import (
	"errors"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
	traQ "github.com/traPtitech/traQ/router/v3"
	"github.com/traPtitech/traQ/utils/random"
)

const traQIssuerName = "traQ"

func (repo *Repository) SyncUsers(info *domain.ConInfo) error {
	if !repo.IsPrevilege(info) {
		return domain.ErrForbidden
	}
	t, err := repo.GormRepo.GetToken(info.ReqUserID)
	if err != nil {
		return defaultErrorHandling(err)
	}
	traQUsers, err := repo.TraQRepo.GetUsers(t, true)
	if err != nil {
		return defaultErrorHandling(err)
	}

	users := make([]*db.User, 0)
	for _, u := range traQUsers {
		if u.Bot {
			continue
		}

		user := &db.User{
			ID:    u.ID,
			State: u.State,
			Provider: db.Provider{
				UserID:  u.ID,
				Issuer:  traQIssuerName,
				Subject: u.ID.String(),
			},
		}
		users = append(users, user)
	}

	err = repo.GormRepo.SyncUsers(users)
	return defaultErrorHandling(err)
}

func (repo *Repository) GetOAuthURL() (url, state, codeVerifier string) {
	return repo.TraQRepo.GetOAuthURL()
}

func (repo *Repository) LoginUser(query, state, codeVerifier string) (*domain.User, error) {
	t, err := repo.TraQRepo.GetOAuthToken(query, state, codeVerifier)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	traQUser, err := repo.TraQRepo.GetUserMe(t)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	user := db.User{
		ID:    traQUser.ID,
		State: 1,
		Token: db.Token{
			UserID: traQUser.ID,
			Token:  t,
		},
		Provider: db.Provider{
			UserID:  traQUser.ID,
			Issuer:  traQIssuerName,
			Subject: traQUser.ID.String(),
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

func (repo *Repository) GetUser(userID uuid.UUID, info *domain.ConInfo) (*domain.User, error) {
	t, err := repo.GormRepo.GetToken(info.ReqUserID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	userMeta, err := repo.GormRepo.GetUser(userID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	if userMeta.Provider.Issuer == traQIssuerName {
		userBody, err := repo.TraQRepo.GetUser(t, userID)
		if err != nil {
			return nil, defaultErrorHandling(err)
		}
		user, _ := repo.mergeUser(userMeta, userBody)
		return user, nil
	}
	// userBody, err := repo.gormRepo.GetUserBody(userID)

	return nil, errors.New("not implemented")
}

func (repo *Repository) GetUserMe(info *domain.ConInfo) (*domain.User, error) {
	return repo.GetUser(info.ReqUserID, info)
}

func (repo *Repository) GetAllUsers(includeSuspend bool, info *domain.ConInfo) ([]*domain.User, error) {
	t, err := repo.GormRepo.GetToken(info.ReqUserID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	userMetas, err := repo.GormRepo.GetAllUsers(!includeSuspend)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	traQUserBodys, err := repo.TraQRepo.GetUsers(t, includeSuspend)
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
		user, _ := repo.mergeUser(userMeta, userBody)
		users = append(users, user)
	}
	return users, nil
}

func (repo *Repository) ReNewMyiCalSecret(info *domain.ConInfo) (secret string, err error) {
	secret = random.SecureAlphaNumeric(16)
	err = repo.GormRepo.UpdateiCalSecret(info.ReqUserID, secret)
	return
}

func (repo *Repository) GetMyiCalSecret(info *domain.ConInfo) (string, error) {
	user, err := repo.GormRepo.GetUser(info.ReqUserID)
	if err != nil {
		return "", defaultErrorHandling(err)
	}
	if user.IcalSecret == "" {
		return "", domain.ErrNotFound
	}
	return user.IcalSecret, nil
}

func (repo *Repository) IsPrevilege(info *domain.ConInfo) bool {
	user, err := repo.GormRepo.GetUser(info.ReqUserID)
	if err != nil {
		return false
	}
	return user.Privilege
}

func traQUserMap(users []*traQ.User) map[uuid.UUID]*traQ.User {
	userMap := make(map[uuid.UUID]*traQ.User)
	for _, user := range users {
		userMap[user.ID] = user
	}
	return userMap
}

func (repo *Repository) mergeUser(userMeta *db.User, userBody *traQ.User) (*domain.User, error) {
	if userMeta.ID != userBody.ID {
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
