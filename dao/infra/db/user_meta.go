package db

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

func saveUser(db *gorm.DB, userID uuid.UUID, privilege bool) (*UserMeta, error) {
	user := UserMeta{
		ID:        userID,
		Privilege: privilege,
	}
	err := db.Create(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
