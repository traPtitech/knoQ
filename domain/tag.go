package domain

import "github.com/gofrs/uuid"

type Tag struct {
	ID   uuid.UUID
	Name string
	Model
}

type TagRepository interface {
	CreateOrGetTag(name string) (*Tag, error)
	GetTagByName(name string) (*Tag, error)
	// UpdateTag(tagID uuid.UUID, name string) (*Tag, error)
	// DeleteTag(tagID uuid.UUID) error
	GetTag(tagID uuid.UUID) (*Tag, error)
	GetAllTags() ([]*Tag, error)
}
