package production

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
)

func (repo *Repository) CreateOrGetTag(name string) (*domain.Tag, error) {
	return repo.GormRepo.CreateOrGetTag(name)
}

func (repo *Repository) GetTag(tagID uuid.UUID) (*domain.Tag, error) {
	return repo.GormRepo.GetTag(tagID)
}

func (repo *Repository) GetAllTags() ([]*domain.Tag, error) {
	return repo.GormRepo.GetAllTags()
}
