package db

import (
	"errors"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

func userPreload(tx *gorm.DB) *gorm.DB {
	return tx
}

func (repo *GormRepository) SaveUser(user User) (*User, error) {
	u, err := saveUser(repo.db, &user)
	return u, defaultErrorHandling(err)
}

func (repo *GormRepository) UpdateiCalSecret(userID uuid.UUID, secret string) error {
	err := updateiCalSecret(repo.db, userID, secret)
	return defaultErrorHandling(err)
}

func (repo *GormRepository) GetUser(userID uuid.UUID) (*User, error) {
	u, err := getUser(userPreload(repo.db), userID)
	return u, defaultErrorHandling(err)
}

func (repo *GormRepository) GetAllUsers(onlyActive bool) ([]*User, error) {
	us, err := getAllUsers(userPreload(repo.db), onlyActive)
	return us, defaultErrorHandling(err)
}

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

	return defaultErrorHandling(err)
}

// saveUser user.IcalSecret == "" の時、値は更新されません。
// また、user.Provider, user.Tokenは空の時、更新されません。
// user.Privilegeは常に更新されません。
func saveUser(db *gorm.DB, user *User) (*User, error) {
	err := db.Transaction(func(tx *gorm.DB) error {
		existingUser, err := getUser(tx.Preload("Token"), user.ID)
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
