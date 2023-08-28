package repository

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
)

func (repo *Repository) CreateOrGetTag(name string) (*domain.Tag, error) {
	t, err := repo.GormRepo.CreateOrGetTag(name)
	return t, defaultErrorHandling(err)
}

func (repo *Repository) GetTag(tagID uuid.UUID) (*domain.Tag, error) {
	t, err := repo.GormRepo.GetTag(tagID)
	return t, defaultErrorHandling(err)
}

func (repo *Repository) GetAllTags() ([]*domain.Tag, error) {
	ts, err := repo.GormRepo.GetAllTags()
	return ts, defaultErrorHandling(err)
}
