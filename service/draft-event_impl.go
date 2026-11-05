package service

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filters"
)

func (s *service) CreateDraftEvent(ctx context.Context, requesterID uuid.UUID, draftEventParams domain.WriteDraftEventParams) (*domain.DraftEvent, error) {
	if !draftEventParams.TimeConsistency() {
		return nil, ErrTimeConsistency
	}

	var draftEventResp *domain.DraftEvent
	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		p := domain.UpsertDraftEventArgs{
			WriteDraftEventParams: draftEventParams,
			CreatedBy:             requesterID,
		}
		var err error
		draftEventResp, err = s.GormRepo.CreateDraftEvent(ctx, p)
		if err != nil {
			return err
		}

		w := make([]domain.WriteTimeWindowAvailability, len(draftEventParams.TimeWindows))
		for i, wd := range draftEventParams.TimeWindows {
			w[i] = domain.WriteTimeWindowAvailability{
				TimeWindowID: wd.TimeWindowID,
				Status:       domain.Pending,
			}
		}

		for _, invitee := range draftEventResp.Invitees {
			s.GormRepo.UpsertDraftEventSchedule(ctx, draftEventResp.ID, invitee.ID, w)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return s.GetDraftEvent(ctx, draftEventResp.ID)
}

func (s *service) UpdateDraftEvent(ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID, draftEventParams domain.WriteDraftEventParams) (*domain.DraftEvent, error) {
	if !s.IsDraftEventAdmins(ctx, requesterID, draftEventID) {
		return nil, domain.ErrForbidden
	}

	if !draftEventParams.TimeConsistency() {
		return nil, ErrTimeConsistency
	}

	var err error
	var draftEventResp *domain.DraftEvent
	err = s.TxManager.Do(ctx, func(ctx context.Context) error {
		p := domain.UpsertDraftEventArgs{
			WriteDraftEventParams: draftEventParams,
			CreatedBy:             requesterID,
		}

		var err error
		draftEventResp, err = s.GormRepo.UpdateDraftEvent(ctx, draftEventID, p)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GetDraftEvent(ctx, draftEventResp.ID)
}

func (s *service) DeleteDraftEvent(ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID) error {
	if !s.IsDraftEventAdmins(ctx, requesterID, draftEventID) {
		return domain.ErrForbidden
	}

	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		return s.GormRepo.DeleteDraftEvent(ctx, draftEventID)
	})
	return err
}

func (s *service) AddDraftEventTag(ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID, tagName string, locked bool) error {
	if locked && !s.IsDraftEventAdmins(ctx, requesterID, draftEventID) {
		return domain.ErrForbidden
	}

	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		return s.GormRepo.AddEventTag(ctx, draftEventID, domain.EventTagParams{
			Name: tagName, Locked: locked,
		})
	})
	return err
}

func (s *service) DeleteDraftEventTag(ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID, tagName string) error {
	deleteLocked := s.IsDraftEventAdmins(ctx, requesterID, draftEventID)

	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		return s.GormRepo.DeleteEventTag(ctx, draftEventID, tagName, deleteLocked)
	})
	return err
}

func (s *service) GetDraftEvent(ctx context.Context, draftEventID uuid.UUID) (*domain.DraftEvent, error) {
	draftEvent, err := s.GormRepo.GetDraftEvent(ctx, draftEventID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	return draftEvent, nil
}

func (s *service) GetDraftEvents(ctx context.Context, requesterID uuid.UUID, expr filters.Expr) ([]*domain.DraftEvent, error) {
	des, err := s.GormRepo.GetAllDraftEvents(ctx, expr)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	return des, nil
}

func (s *service) UpsertMeDraftEventAvailability(ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID, schedule []domain.WriteTimeWindowAvailability) error {
	draftEvent, err := s.GetEvent(ctx, draftEventID)
	if err != nil {
		return err
	}
	if !draftEvent.Open {
		return domain.ErrForbidden
	}

	err = s.TxManager.Do(ctx, func(ctx context.Context) error {
		return s.GormRepo.UpsertDraftEventSchedule(ctx, draftEventID, requesterID, schedule)
	})
	if err != nil {
		return defaultErrorHandling(err)
	}
	return nil
}

func (s *service) IsDraftEventAdmins(ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID) bool {
	draftEvent, err := s.GormRepo.GetDraftEvent(ctx, draftEventID)
	if err != nil {
		return false
	}
	for _, admin := range draftEvent.Admins {
		if requesterID == admin.ID {
			return true
		}
	}
	return false
}
