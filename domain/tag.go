package domain

import (
	"context"

	"github.com/gofrs/uuid"
)

type Tag struct {
	ID   uuid.UUID
	Name string
	Model
}

type TagService interface {
	CreateOrGetTag(ctx context.Context, name string) (*Tag, error)
	GetTag(ctx context.Context, tagID uuid.UUID) (*Tag, error)
	GetAllTags(ctx context.Context) ([]*Tag, error)
}

type TagRepository interface {
	CreateOrGetTag(name string) (*Tag, error)
	GetTag(tagID uuid.UUID) (*Tag, error)
	GetAllTags() ([]*Tag, error)
}
