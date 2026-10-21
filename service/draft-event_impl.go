package service

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filters"
)

func (s *service) CreateDraftEvent(ctx context.Context, requesterID uuid.UUID, draftEventParams domain.WriteDraftEventParams) (*domain.DraftEvent, error) {
	return nil, nil
}

func (s *service) UpdateDraftEvent(ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID, draftEventParams domain.WriteDraftEventParams) (*domain.DraftEvent, error) {
	return nil, nil
}

func (s *service) DeleteDraftEvent(ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID) error {
	return nil
}

func (s *service) AddDraftEventTag(ctx context.Context, reqesterID uuid.UUID, draftEventID uuid.UUID, tagName string, locked bool) error {
	return nil
}

func (s *service) DeleteDraftEventTag(ctx context.Context, reqID uuid.UUID, draftEventID uuid.UUID) error {
	return nil
}

func (s *service) GetDraftEvent(ctx context.Context, draftEventID uuid.UUID) (*domain.DraftEvent, error) {
	return nil, nil
}

func (s *service) GetDraftEvents(ctx context.Context, reqesterID uuid.UUID, expr filters.Expr) ([]*domain.DraftEvent, error) {
	return []*domain.DraftEvent{}, nil
}

func (s *service) UpsertMeDraftEventAvailability(ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID, schedulde []domain.WriteTimeWindowAvailability) error {
	return nil
}

func (s *service) IsDraftEventAdmins(ctx context.Context, reqesterID uuid.UUID, draftEventID uuid.UUID) bool {
	return false
}
