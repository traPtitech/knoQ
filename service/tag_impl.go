package service

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
)

func (s *service) CreateOrGetTag(ctx context.Context, name string) (*domain.Tag, error) {
	t, err := s.GormRepo.CreateOrGetTag(name)
	return t, defaultErrorHandling(err)
}

func (s *service) GetTag(ctx context.Context, tagID uuid.UUID) (*domain.Tag, error) {
	t, err := s.GormRepo.GetTag(tagID)
	return t, defaultErrorHandling(err)
}

func (s *service) GetAllTags(ctx context.Context) ([]*domain.Tag, error) {
	ts, err := s.GormRepo.GetAllTags()
	return ts, defaultErrorHandling(err)
}
