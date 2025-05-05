package db

import (
	"errors"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra"
	"gorm.io/gorm"
)

func mergeDBUserandTraQUser(dbUser *User, traqUser *infra.TraqUserResponse) (*domain.User, error) {
	if dbUser.ID != traqUser.ID {
		return nil, errors.New("id does not match")
	}
	if dbUser.ProviderName != "traQ" {
		return nil, errors.New("different provider")
	}
	return &domain.User{
		ID:          dbUser.ID,
		Name:        traqUser.Name,
		DisplayName: traqUser.DisplayName,
		Icon:        traqUser.IconURL,
		Privileged:  dbUser.Privileged,
		State:       dbUser.State,
	}, nil
}

func traQUserMap(users []*infra.TraqUserResponse) map[uuid.UUID]*infra.TraqUserResponse {
	userMap := make(map[uuid.UUID]*infra.TraqUserResponse)
	for _, u := range users {
		userMap[u.ID] = u
	}
	return userMap
}

func (repo *GormRepository) SaveUser(user User) (*domain.User, error) {
	u, err := saveUser(repo.db, &user)
	if err != nil {
		return nil, err
	}

	tu, err := repo.traqRepo.GetUser(u.ID)
	if err != nil {
		return nil, err
	}

	du, err := mergeDBUserandTraQUser(u, tu)

	return du, defaultErrorHandling(err)
}

func (repo *GormRepository) UpdateiCalSecret(userID uuid.UUID, secret string) error {
	err := updateiCalSecret(repo.db, userID, secret)
	return defaultErrorHandling(err)
}

func (repo *GormRepository) GetUser(userID uuid.UUID) (*domain.User, error) {
	u, err := getUser(repo.db, userID)
	if err != nil {
		return nil, err
	}

	tu, err := repo.traqRepo.GetUser(u.ID)
	if err != nil {
		return nil, err
	}

	du, err := mergeDBUserandTraQUser(u, tu)

	return du, defaultErrorHandling(err)
}

func (repo *GormRepository) GetAllUsers(onlyActive bool) ([]*domain.User, error) {
	us, err := getAllUsers(repo.db, onlyActive)

	traQUsers, err := repo.traqRepo.GetUsers(true)
	if err != nil {
		return nil, err
	}

	traQUserBodsMap := traQUserMap(traQUsers)
	users := make([]*domain.User, 0, len(us))
	for _, u := range us {
		traqUser, ok := traQUserBodsMap[u.ID]
		if !ok {
			continue
		}
		user, err := mergeDBUserandTraQUser(u, traqUser)
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, defaultErrorHandling(err)
}

func (repo *GormRepository) SyncUsers() error {
	traQUsers, err := repo.traqRepo.GetUsers(true)
	if err != nil {
		return err
	}
	users := make([]*User, 0)
	for _, u := range traQUsers {
		if u.Bot {
			continue
		}

		if u.ID.IsNil() {
			panic("uuid is nil")
		}
		user := &User{
			ID:    u.ID,
			State: int(u.State),
			// TODO: traQIssuerName の変数
			ProviderName: "traQ",
		}
		users = append(users, user)
	}

	err = repo.db.Transaction(func(tx *gorm.DB) error {
		existingUsers, err := getAllUsers(tx, false)
		if err != nil {
			return err
		}

		for _, u := range users {
			exist := false
			for _, eu := range existingUsers {
				if u.ID == eu.ID {
					// contains
					exist = true
					if u.State != eu.State {
						eu.State = u.State
						_, err := saveUser(tx, eu)
						if err != nil {
							return err
						}
					}
					break
				}
			}

			if !exist {
				_, err := saveUser(tx, u)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	return defaultErrorHandling(err)
}

func (repo *GormRepository) GetiCalSecret(userID uuid.UUID) (string, error) {
	u, err := getUser(repo.db, userID)
	if err != nil {
		return "", err
	}

	return u.IcalSecret, nil
}

// saveUser user.IcalSecret == "" の時、値は更新されません。
// また、user.Provider, user.Tokenは空の時、更新されません。
// user.Privilegeは常に更新されません。
func saveUser(db *gorm.DB, user *User) (*User, error) {
	err := db.Transaction(func(tx *gorm.DB) error {
		existingUser, err := getUser(tx, user.ID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&user).Error
		}
		if err != nil {
			return err
		}

		if user.IcalSecret == "" {
			user.IcalSecret = existingUser.IcalSecret
		}
		return tx.Session(&gorm.Session{FullSaveAssociations: true}).Updates(user).Error
	})

	return user, err
}

func updateiCalSecret(db *gorm.DB, userID uuid.UUID, secret string) error {
	return db.Model(&User{ID: userID}).Update("ical_secret", secret).Error
}

func getUser(db *gorm.DB, userID uuid.UUID) (*User, error) {
	user := User{}
	err := db.Take(&user, userID).Error
	return &user, err
}

func getAllUsers(db *gorm.DB, onlyActive bool) ([]*User, error) {
	users := make([]*User, 0)
	if onlyActive {
		err := db.Where("state = ?", 1).Find(&users).Error
		return users, err
	}
	err := db.Find(&users).Error
	return users, err
}

func (repo *GormRepository) GrantPrivilege(userID uuid.UUID) error {
	err := grantPrivilege(repo.db, userID)
	return defaultErrorHandling(err)
}

func grantPrivilege(db *gorm.DB, userID uuid.UUID) error {
	err := db.Model(&User{ID: userID}).Update("privilege", true).Error
	return err
}
