package repository

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
	"github.com/traPtitech/knoQ/utils/random"
)

const traQIssuerName = "traQ"

func (repo *Repository) SyncUsers(info *domain.ConInfo) error {
	if !repo.IsPrivilege(info) {
		return domain.ErrForbidden
	}
	// traQUsers, err := repo.TraQRepo.GetUsers(true)
	// if err != nil {
	// 	return err
	// }

	// users := make([]*db.User, 0)
	// for _, u := range traQUsers {
	// 	if u.Bot {
	// 		continue
	// 	}

	// 	// uid := uuid.Must(uuid.FromString(u.GetId()))
	// 	if u.ID.IsNil() {
	// 		panic("uuid is nil")
	// 	}
	// 	user := &db.User{
	// 		ID:           u.ID,
	// 		State:        int(u.State),
	// 		ProviderName: traQIssuerName,
	// 	}
	// 	users = append(users, user)
	// }

	err := repo.GormRepo.SyncUsers()
	return err
}

func (repo *Repository) GetOAuthURL() (url, state, codeVerifier string) {
	return repo.GormRepo.GetTraqRepository().GetOAuthURL()
}

func (repo *Repository) LoginUser(query, state, codeVerifier string) (*domain.User, error) {
	t, err := repo.GormRepo.GetTraqRepository().GetOAuthToken(query, state, codeVerifier)
	if err != nil {
		return nil, err
	}
	traQUser, err := repo.GormRepo.GetTraqRepository().GetUserMe(t.AccessToken)
	if err != nil {
		return nil, err
	}
	// uid := uuid.Must(uuid.FromString(traQUser.GetId()))
	if traQUser.ID.IsNil() {
		panic("uuid is nil")
	}
	user := db.User{
		ID:           traQUser.ID,
		State:        1,
		AccessToken:  t.AccessToken,
		ProviderName: traQIssuerName,
	}
	_, err = repo.GormRepo.SaveUser(user)
	if err != nil {
		println("hererererere")
		return nil, err
	}
	u, err := repo.GetUser(user.ID, &domain.ConInfo{
		ReqUserID: user.ID,
	})
	return u, err
}

func (repo *Repository) GetUser(userID uuid.UUID, _ *domain.ConInfo) (*domain.User, error) {
	user, err := repo.GormRepo.GetUser(userID)
	if err != nil {
		return nil, err
	}

	// if dbUser.ProviderName == traQIssuerName {
	// 	traqUser, err := repo.GormRepo.GetTraqRepository().GetUser(userID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	user, _ := repo.mergeDBUserandTraQUser(dbUser, traqUser)
	return user, nil
	// }
}

func (repo *Repository) GetUserMe(info *domain.ConInfo) (*domain.User, error) {
	return repo.GetUser(info.ReqUserID, info)
}

func (repo *Repository) GetAllUsers(includeSuspend, includeBot bool, _ *domain.ConInfo) ([]*domain.User, error) {
	users, err := repo.GormRepo.GetAllUsers(!includeSuspend)
	if err != nil {
		return nil, err
	}
	// TODO fix
	// traQUserBodys, err := repo.GormRepo.GetTraqRepository().GetUsers(includeSuspend)
	// if err != nil {
	// 	return nil, err
	// }
	// traQUserBodsMap := traQUserMap(traQUserBodys)
	// users := make([]*domain.User, 0, len(dbUsers))
	// for _, dbUser := range dbUsers {
	// 	traqUser, ok := traQUserBodsMap[dbUser.ID]
	// 	if !ok {
	// 		continue
	// 	}
	// 	if !includeBot && traqUser.Bot {
	// 		continue
	// 	}
	// 	user, err := repo.mergeDBUserandTraQUser(dbUser, traqUser)
	// 	if err != nil {
	// 		continue
	// 	}
	// 	users = append(users, user)
	// }

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
		return "", err
	}
	if user.State != 1 {
		return "", domain.ErrForbidden
	}

	icalSecret, err := repo.GormRepo.GetiCalSecret(info.ReqUserID)

	if icalSecret == "" {
		return "", domain.ErrNotFound
	}
	return icalSecret, nil
}

func (repo *Repository) IsPrivilege(info *domain.ConInfo) bool {
	user, err := repo.GormRepo.GetUser(info.ReqUserID)
	if err != nil {
		return false
	}
	return user.Privileged
}

func (repo *Repository) GrantPrivilege(userID uuid.UUID) error {
	user, err := repo.GormRepo.GetUser(userID)
	if err != nil {
		return err
	}
	if user.Privileged {
		return fmt.Errorf("%w: user has been already privileged", domain.ErrBadRequest)
	}
	err = repo.GormRepo.GrantPrivilege(userID)
	return err
}
