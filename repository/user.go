package repository

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

func (repo *Repository) SyncUsers(info *domain.ConInfo) error {
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
			ID:           uid,
			State:        int(u.State),
			ProviderName: traQIssuerName,
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
	traQUser, err := repo.TraQRepo.GetUserMe(t.AccessToken)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	uid := uuid.Must(uuid.FromString(traQUser.GetId()))
	user := db.User{
		ID:           uid,
		State:        1,
		AccessToken:  t.AccessToken,
		ProviderName: traQIssuerName,
	}
	_, err = repo.GormRepo.SaveUser(user)
	if err != nil {
		println("hererererere")
		return nil, defaultErrorHandling(err)
	}
	u, err := repo.GetUser(user.ID, &domain.ConInfo{
		ReqUserID: user.ID,
	})
	return u, defaultErrorHandling(err)
}

func (repo *Repository) GetUser(userID uuid.UUID, _ *domain.ConInfo) (*domain.User, error) {
	dbUser, err := repo.GormRepo.GetUser(userID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	if dbUser.ProviderName == traQIssuerName {
		traqUser, err := repo.TraQRepo.GetUser(userID)
		if err != nil {
			return nil, defaultErrorHandling(err)
		}
		user, _ := repo.mergeDBUserandTraQUser(dbUser, traqUser)
		return user, nil
	}

	return nil, errors.New("not implemented")
}

func (repo *Repository) GetUserMe(info *domain.ConInfo) (*domain.User, error) {
	return repo.GetUser(info.ReqUserID, info)
}

func (repo *Repository) GetAllUsers(includeSuspend, includeBot bool, _ *domain.ConInfo) ([]*domain.User, error) {
	dbUsers, err := repo.GormRepo.GetAllUsers(!includeSuspend)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	// TODO fix
	traQUserBodys, err := repo.TraQRepo.GetUsers(includeSuspend)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	traQUserBodsMap := traQUserMap(traQUserBodys)
	users := make([]*domain.User, 0, len(dbUsers))
	for _, dbUser := range dbUsers {
		traqUser, ok := traQUserBodsMap[dbUser.ID]
		if !ok {
			continue
		}
		if !includeBot && traqUser.Bot {
			continue
		}
		user, err := repo.mergeDBUserandTraQUser(dbUser, traqUser)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

func (repo *Repository) ReNewMyiCalSecret(info *domain.ConInfo) (secret string, err error) {
	secret = random.AlphaNumeric(16, true)
	err = repo.GormRepo.UpdateiCalSecret(info.ReqUserID, secret)
	return
}

func (repo *Repository) GetMyiCalSecret(info *domain.ConInfo) (string, error) {
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

func (repo *Repository) IsPrivilege(info *domain.ConInfo) bool {
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

func (repo *Repository) mergeDBUserandTraQUser(dbUser *db.User, traqUser *traq.User) (*domain.User, error) {
	if dbUser.ID != uuid.Must(uuid.FromString(traqUser.GetId())) {
		return nil, errors.New("id does not match")
	}
	if dbUser.ProviderName != traQIssuerName {
		return nil, errors.New("different provider")
	}
	return &domain.User{
		ID:          dbUser.ID,
		Name:        traqUser.Name,
		DisplayName: traqUser.DisplayName,
		Icon:        repo.TraQRepo.URL + "/public/icon/" + traqUser.Name,
		Privileged:  dbUser.Privilege,
		State:       dbUser.State,
	}, nil
}

func (repo *Repository) GrantPrivilege(userID uuid.UUID) error {
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
