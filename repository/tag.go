package repository

import (
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
)

type TagRepository interface {
	CreateOrGetTag(name string) (*Tag, error)
	GetTagByName(name string) (*Tag, error)
	// UpdateTag(tagID uuid.UUID, name string) (*Tag, error)
	// DeleteTag(tagID uuid.UUID) error
	GetTag(tagID uuid.UUID) (*Tag, error)
	GetAllTags() ([]*Tag, error)
}

func (repo *GormRepository) CreateOrGetTag(name string) (*Tag, error) {
	tag := new(Tag)

	if err := repo.DB.Where("name = ?", name).Take(&tag).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		tag.Name = name
		err = repo.DB.Create(&tag).Error
		if err != nil {
			return nil, err
		}
	}
	return tag, nil
}

func (repo *GormRepository) GetTagByName(name string) (*Tag, error) {
	tag := new(Tag)
	err := repo.DB.Where("name = ?", name).Take(&tag).Error
	return tag, err
}

func (repo *GormRepository) GetTag(tagID uuid.UUID) (*Tag, error) {
	tag := new(Tag)
	tag.ID = tagID
	err := repo.DB.Take(&tag).Error
	return tag, err
}

func (repo *GormRepository) GetAllTags() ([]*Tag, error) {
	tags := make([]*Tag, 0)
	err := repo.DB.Find(&tags).Error
	return tags, err
}

// BeforeCreate is gorm hook
func (t *Tag) BeforeCreate() (err error) {
	t.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}
