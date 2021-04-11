package production

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
)

func (repo *Repository) CreateOrGetTag(name string) (*domain.Tag, error) {
	return repo.gormRepo.CreateOrGetTag(name)
}

func (repo *Repository) GetTag(tagID uuid.UUID) (*domain.Tag, error) {
	return repo.gormRepo.GetTag(tagID)
}

func (repo *Repository) GetAllTags() ([]*domain.Tag, error) {
	return repo.gormRepo.GetAllTags()
}
