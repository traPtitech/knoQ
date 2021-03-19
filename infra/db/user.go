package db

import (
	"errors"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// saveUser user.IcalSecret == "" の時、値は更新されません。
// また、user.Provider, user.Tokenは空の時、更新されません。
// user.Previlegeは常に更新されません。
func saveUser(db *gorm.DB, user *User) (*User, error) {
	err := db.Transaction(func(tx *gorm.DB) error {
		existingUser, err := getUser(tx.Preload("Provider").Preload("Token"), user.ID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&user).Error
		}
		if err != nil {
			return err
		}

		if user.IcalSecret == "" {
			user.IcalSecret = existingUser.IcalSecret
		}
		return tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(&user).Error
	})

	return user, err
}

func (repo *GormRepository) SaveUser(user User) (*User, error) {
	return saveUser(repo.db, &user)
}

func (repo *GormRepository) Privilege(userID uuid.UUID) bool {
	user, err := getUser(repo.db, userID)
	if err != nil {
		return false
	}
	return user.Privilege
}

func getUser(db *gorm.DB, userID uuid.UUID) (*User, error) {
	user := User{
		ID: userID,
	}
	err := db.Take(&user).Error
	return &user, err
}

func (repo *GormRepository) GetUser(userID uuid.UUID) (*User, error) {
	return getUser(repo.db.Preload("Provider"), userID)
}

func (repo *GormRepository) GetAllUsers(onlyActive bool) ([]*User, error) {
	return getAllUsers(repo.db, onlyActive)
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

// func (repo *GormRepository) UpdateUserState(userID uuid.UUID, state int) error {
// user := User{
// ID: userID,
// }
// return repo.db.Model(&user).Update("state", state).Error
// }

func (repo *GormRepository) SyncUsers(users []*User) error {
	err := repo.db.Transaction(func(tx *gorm.DB) error {
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

	return err
}
