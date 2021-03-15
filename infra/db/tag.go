package db

import (
	"gorm.io/gorm"
)

func createTag(db *gorm.DB, name string) (*Tag, error) {
	tag := Tag{Name: name}
	err := db.Create(&tag).Error
	return &tag, err
}
