package db

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

func saveUser(db *gorm.DB, user *User) (*User, error) {
	err := db.Save(&user).Error
	return user, err
}

func (repo *GormRepository) SaveUser(user User) (*User, error) {
	err := repo.db.Save(&user).Error
	return &user, err
}

func (repo *GormRepository) Privilege(userID uuid.UUID) bool {
	user, err := getUser(repo.db, userID)
	if err != nil {
		return false
	}
	return user.Privilege
}

func getUser(db *gorm.DB, userID uuid.UUID) (*User, error) {
	var user User
	err := db.Take(&user).Error
	return &user, err
}

func (repo *GormRepository) GetAllUsers() (users []*User, err error) {
	err = repo.db.Find(&users).Error
	return
}

func getAllUsers(db *gorm.DB) (users []*User, err error) {
	err = db.Find(&users).Error
	return
}

// func (repo *GormRepository) UpdateUserState(userID uuid.UUID, state int) error {
// user := User{
// ID: userID,
// }
// return repo.db.Model(&user).Update("state", state).Error
// }

func (repo *GormRepository) SyncUsers(users []*User) error {
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		existingUsers, err := getAllUsers(tx)
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
