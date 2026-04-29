package service

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
)

func (s *service) CreateUnVerifiedRoom(ctx context.Context, reqID uuid.UUID, params domain.WriteRoomParams) (*domain.Room, error) {
	if !params.TimeConsistency() {
		return nil, ErrTimeConsistency
	}
	p := domain.CreateRoomArgs{
		WriteRoomParams: params,
		Verified:        false,
		CreatedBy:       reqID,
	}
	var roomResp *domain.Room
	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		var err error
		roomResp, err = s.GormRepo.CreateRoom(ctx, p)
		return err
	})
	return roomResp, defaultErrorHandling(err)
}

func (s *service) CreateVerifiedRoom(ctx context.Context, reqID uuid.UUID, params domain.WriteRoomParams) (*domain.Room, error) {

	if !s.IsPrivilege(ctx, reqID) {
		return nil, domain.ErrForbidden
	}
	if !params.TimeConsistency() {
		return nil, ErrTimeConsistency
	}
	p := domain.CreateRoomArgs{
		WriteRoomParams: params,
		Verified:        true,
		CreatedBy:       reqID,
	}

	var roomResp *domain.Room
	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		var err error
		roomResp, err = s.GormRepo.CreateRoom(ctx, p)
		return err
	})
	return roomResp, defaultErrorHandling(err)
}

func (s *service) UpdateRoom(ctx context.Context, reqID uuid.UUID, roomID uuid.UUID, params domain.WriteRoomParams) (*domain.Room, error) {
	if roomID == uuid.Nil {
		return nil, ErrRoomUndefined
	}
	if !s.IsRoomAdmins(ctx, reqID, roomID) {
		return nil, domain.ErrForbidden
	}

	p := domain.UpdateRoomArgs{
		WriteRoomParams: params,
		CreatedBy:       reqID,
	}

	if !params.TimeConsistency() {
		return nil, ErrTimeConsistency
	}

	var roomResp *domain.Room
	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		var err error
		roomResp, err = s.GormRepo.UpdateRoom(ctx, roomID, p)
		return err
	})
	return roomResp, defaultErrorHandling(err)
}

func (s *service) VerifyRoom(ctx context.Context, reqID uuid.UUID, roomID uuid.UUID) error {
	if !s.IsPrivilege(ctx, reqID) {
		return domain.ErrForbidden
	}

	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		err := s.GormRepo.UpdateRoomVerified(ctx, roomID, true)
		return err
	})

	return defaultErrorHandling(err)
}

func (s *service) UnVerifyRoom(ctx context.Context, reqID uuid.UUID, roomID uuid.UUID) error {
	if !s.IsPrivilege(ctx, reqID) {
		return domain.ErrForbidden
	}
	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		err := s.GormRepo.UpdateRoomVerified(ctx, roomID, false)
		return err
	})

	return defaultErrorHandling(err)
}

func (s *service) DeleteRoom(ctx context.Context, reqID uuid.UUID, roomID uuid.UUID) error {
	if !s.IsRoomAdmins(ctx, reqID, roomID) {
		return domain.ErrForbidden
	}
	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		err := s.GormRepo.DeleteRoom(ctx, roomID)
		return err
	})

	return defaultErrorHandling(err)
}

func (s *service) GetRoom(ctx context.Context, roomID uuid.UUID, excludeEventID uuid.UUID) (*domain.Room, error) {
	rs, err := s.GormRepo.GetRoom(ctx, roomID, excludeEventID)
	return rs, defaultErrorHandling(err)
}

func (s *service) GetAllRooms(ctx context.Context, start time.Time, end time.Time, excludeEventID uuid.UUID) ([]*domain.Room, error) {
	rs, err := s.GormRepo.GetAllRooms(ctx, start, end, excludeEventID)
	return rs, defaultErrorHandling(err)
}

func (s *service) IsRoomAdmins(ctx context.Context, reqID uuid.UUID, roomID uuid.UUID) bool {
	room, err := s.GetRoom(ctx, roomID, uuid.Nil)
	if err != nil {
		return false
	}
	for _, admin := range room.Admins {
		if reqID == admin.ID {
			return true
		}
	}
	return false
}
