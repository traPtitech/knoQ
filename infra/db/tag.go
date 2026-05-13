package db

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

func (repo *gormRepository) CreateOrGetTag(ctx context.Context, name string) (*domain.Tag, error) {
	tag, err := createOrGetTag(getTx(ctx, repo.db.WithContext(ctx)), name)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	t := ConvTagTodomainTag(*tag)
	return &t, nil
}

func (repo *gormRepository) GetTag(ctx context.Context, tagID uuid.UUID) (*domain.Tag, error) {
	tag, err := getTag(getTx(ctx, repo.db.WithContext(ctx)), tagID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	t := ConvTagTodomainTag(*tag)
	return &t, nil
}

func (repo *gormRepository) GetAllTags(ctx context.Context) ([]*domain.Tag, error) {
	tags, err := getAllTags(getTx(ctx, repo.db.WithContext(ctx)))
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	t := ConvSPTagToSPdomainTag(tags)
	return t, nil
}

func createOrGetTag(db *gorm.DB, name string) (*Tag, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	tag := Tag{Name: name}
	err = db.Where(&Tag{Name: name}).Attrs(Tag{ID: id}).FirstOrCreate(&tag).Error
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
