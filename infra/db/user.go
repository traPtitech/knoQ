package db

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

func saveUser(db *gorm.DB, userID uuid.UUID, privilege bool) (*User, error) {
	user := User{
		ID:        userID,
		Privilege: privilege,
	}
	err := db.Create(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *GormRepository) SaveUser(user User) (*User, error) {
	err := repo.db.FirstOrCreate(&user).Error
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
