package db

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// BeforeCreate is hook
func (t *Tag) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}

func createTag(db *gorm.DB, name string) (*Tag, error) {
	tag := Tag{Name: name}
	err := db.Create(&tag).Error
	return &tag, err
}
