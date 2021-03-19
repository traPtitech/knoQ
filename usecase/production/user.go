package production

import (
	"errors"

	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
)

func (repo *Repository) SyncUsers(info *domain.ConInfo) error {
	if !repo.gormRepo.Privilege(info.ReqUserID) {
		return errors.New("fobidden")
	}
	t, err := repo.gormRepo.GetToken(info.ReqUserID)
	if err != nil {
		return err
	}
	traQUsers, err := repo.traQRepo.GetUsers(t, true)
	if err != nil {
		return err
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
				Issuer:  "traQ",
				Subject: u.ID.String(),
			},
		}
		users = append(users, user)
	}

	return repo.gormRepo.SyncUsers(users)
}

func (repo *Repository) LoginUser(query, state, codeVerifier string) error {
	t, err := repo.traQRepo.GetOAuthToken(query, state, codeVerifier)
	if err != nil {
		return err
	}
	traQUser, err := repo.traQRepo.GetUserMe(t)
	if err != nil {
		return err
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
			Issuer:  "traQ",
			Subject: traQUser.ID.String(),
		},
	}
	_, err = repo.gormRepo.SaveUser(user)
	return err
	//if err != nil {
	//return err
	//}

	//user := ConvertPointerv3UserToPointerdomainUser(traQUser)
	//user.Icon = repo.traQRepo.URL + "/public/icon/" + user.Name
	//user.Privileged = userMeta.Privilege
	//user.IsTrap = true

	//return user, nil
}
