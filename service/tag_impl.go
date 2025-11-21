package service

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
)

func (repo *service) CreateOrGetTag(name string) (*domain.Tag, error) {
	t, err := repo.GormRepo.CreateOrGetTag(name)
	return t, defaultErrorHandling(err)
}

func (repo *service) GetTag(tagID uuid.UUID) (*domain.Tag, error) {
	t, err := repo.GormRepo.GetTag(tagID)
	return t, defaultErrorHandling(err)
}

func (repo *service) GetAllTags() ([]*domain.Tag, error) {
	ts, err := repo.GormRepo.GetAllTags()
	return ts, defaultErrorHandling(err)
}
