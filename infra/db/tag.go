package db

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

func (repo *gormRepository) CreateOrGetTag(name string) (*domain.Tag, error) {
	tag, err := createOrGetTag(repo.db, name)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	t := ConvTagTodomainTag(*tag)
	return &t, nil
}

func (repo *gormRepository) GetTag(tagID uuid.UUID) (*domain.Tag, error) {
	tag, err := getTag(repo.db, tagID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	t := ConvTagTodomainTag(*tag)
	return &t, nil
}

func (repo *gormRepository) GetAllTags() ([]*domain.Tag, error) {
	tags, err := getAllTags(repo.db)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	t := ConvSPTagToSPdomainTag(tags)
	return t, nil
}

func createOrGetTag(db *gorm.DB, name string) (*Tag, error) {
	tag := Tag{Name: name}
	err := db.FirstOrCreate(&tag).Error
	return &tag, err
}

func getTag(db *gorm.DB, tagID uuid.UUID) (*Tag, error) {
	tag := Tag{}
	err := db.Take(&tag, tagID).Error
	return &tag, err
}

func getAllTags(db *gorm.DB) ([]*Tag, error) {
	tags := make([]*Tag, 0)
	err := db.Order("name").Find(&tags).Error
	return tags, err
}
