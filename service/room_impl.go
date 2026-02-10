package service

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
)

func (s *service) CreateUnVerifiedRoom(ctx context.Context, reqID uuid.UUID, params domain.WriteRoomParams) (*domain.Room, error) {
	p := domain.CreateRoomArgs{
		WriteRoomParams: params,
		Verified:        false,
		CreatedBy:       reqID,
	}
	r, err := s.GormRepo.CreateRoom(p)
	return r, defaultErrorHandling(err)
}

func (s *service) CreateVerifiedRoom(ctx context.Context, reqID uuid.UUID, params domain.WriteRoomParams) (*domain.Room, error) {

	if !s.IsPrivilege(ctx, reqID) {
		return nil, domain.ErrForbidden
	}
	p := domain.CreateRoomArgs{
		WriteRoomParams: params,
		Verified:        true,
		CreatedBy:       reqID,
	}
	r, err := s.GormRepo.CreateRoom(p)
	return r, defaultErrorHandling(err)
}

func (s *service) UpdateRoom(ctx context.Context, reqID uuid.UUID, roomID uuid.UUID, params domain.WriteRoomParams) (*domain.Room, error) {
	if !s.IsRoomAdmins(ctx, reqID, roomID) {
		return nil, domain.ErrForbidden
	}
	// そもそもNilだとIsRoomAdminsで弾かれるはず
	// if roomID  == uuid.Nil {
	// 	return nil,ErrRoomUndefined
	// }

	p := domain.UpdateRoomArgs{
		WriteRoomParams: params,
		CreatedBy:       reqID,
	}

	// 時間整合性の確認
	if(!params.TimeStart.Before(params.TimeEnd)){
		return nil,ErrTimeConsistency
	}
	r, err := s.GormRepo.UpdateRoom(roomID, p)
	return r, defaultErrorHandling(err)
}

func (s *service) VerifyRoom(ctx context.Context, reqID uuid.UUID, roomID uuid.UUID) error {
	if !s.IsPrivilege(ctx, reqID) {
		return domain.ErrForbidden
	}
	err := s.GormRepo.UpdateRoomVerified(roomID, true)
	return defaultErrorHandling(err)
}

func (s *service) UnVerifyRoom(ctx context.Context, reqID uuid.UUID, roomID uuid.UUID) error {
	if !s.IsPrivilege(ctx, reqID) {
		return domain.ErrForbidden
	}
	err := s.GormRepo.UpdateRoomVerified(roomID, false)
	return defaultErrorHandling(err)
}

func (s *service) DeleteRoom(ctx context.Context, reqID uuid.UUID, roomID uuid.UUID) error {
	if !s.IsRoomAdmins(ctx, reqID, roomID) {
		return domain.ErrForbidden
	}
	err := s.GormRepo.DeleteRoom(roomID)
	return defaultErrorHandling(err)
}

func (s *service) GetRoom(ctx context.Context, roomID uuid.UUID, excludeEventID uuid.UUID) (*domain.Room, error) {
	rs, err := s.GormRepo.GetRoom(roomID, excludeEventID)
	return rs, defaultErrorHandling(err)
}

func (s *service) GetAllRooms(ctx context.Context, start time.Time, end time.Time, excludeEventID uuid.UUID) ([]*domain.Room, error) {
	rs, err := s.GormRepo.GetAllRooms(start, end, excludeEventID)
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
