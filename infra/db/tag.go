package db

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

func createOrGetTag(db *gorm.DB, name string) (*Tag, error) {
	tag := Tag{Name: name}
	err := db.FirstOrCreate(&tag).Error
	return &tag, err
}

func getTag(db *gorm.DB, tagID uuid.UUID) (*Tag, error) {
	tag := Tag{
		ID: tagID,
	}
	err := db.Take(&tag).Error
	return &tag, err
}

func getAllTags(db *gorm.DB) ([]*Tag, error) {
	tags := make([]*Tag, 0)
	err := db.Order("name").Find(&tags).Error
	return tags, err
}
