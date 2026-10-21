package db

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filters"
	"gorm.io/gorm"
)

func draftEventFullPreload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Group").Preload("Group.Members").Preload("Group.Admins").Preload("Group.CreatedBy").
		Preload("Room").Preload("Room.Events").Preload("Room.Admins").Preload("Room.CreatedBy").
		Preload("Admins").Preload("Admins.User").
		Preload("Tags").Preload("Tags.Tag").
		Preload("Attendees").Preload("Attendees.User").
		Preload("CreatedBy")
}

func (repo *gormRepository) CreateDraftEvent(context.Context, domain.UpsertDraftEventArgs) (*domain.DraftEvent, error) {
	return nil, nil
}

func (repo *gormRepository) UpdateDraftEvent(ctx context.Context, draftEventID uuid.UUID, args domain.UpsertDraftEventArgs) (*domain.DraftEvent, error) {
	return nil, nil
}

func (repo *gormRepository) DeleteDraftEvent(ctx context.Context, draftEventID uuid.UUID) error {
	return nil
}

func (repo *gormRepository) UpsertDraftEventSchedule(ctx context.Context, drraftEventID uuid.UUID, userID uuid.UUID, ScheduleStatus []domain.WriteTimeWindowAvailability) error {
	return nil
}

func (repo *gormRepository) GetDraftEvent(ctx context.Context, draftEventID uuid.UUID) (*domain.DraftEvent, error) {
	return nil, nil
}

func (repo *gormRepository) GetAllDraftEvents(ctx context.Context, expr filters.Expr) ([]*domain.DraftEvent, error) {
	return []*domain.DraftEvent{}, nil
}
