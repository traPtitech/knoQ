package db

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// BeforeCreate is hook
func (um *UserMeta) BeforeCreate(tx *gorm.DB) (err error) {
	um.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}

func saveUser(db *gorm.DB, userID uuid.UUID, isAdmin, istraQ bool) (*UserMeta, error) {
	user := UserMeta{
		ID:     userID,
		Admin:  isAdmin,
		IsTraq: istraQ,
	}
	err := db.Create(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
