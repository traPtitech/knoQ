package service

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
)

func (s *service) CreateOrGetTag(ctx context.Context, name string) (*domain.Tag, error) {
	var t *domain.Tag
	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		var err error
		t, err = s.GormRepo.CreateOrGetTag(ctx, name)
		return err
	})
	return t, defaultErrorHandling(err)
}

func (s *service) GetTag(ctx context.Context, tagID uuid.UUID) (*domain.Tag, error) {
	t, err := s.GormRepo.GetTag(ctx, tagID)
	return t, defaultErrorHandling(err)
}

func (s *service) GetAllTags(ctx context.Context) ([]*domain.Tag, error) {
	ts, err := s.GormRepo.GetAllTags(ctx)
	return ts, defaultErrorHandling(err)
}
