package service

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
)

func (repo *service) CreateUnVerifiedRoom(ctx context.Context, params domain.WriteRoomParams) (*domain.Room, error) {
	reqID, _ := domain.GetUserID(ctx)

	p := db.CreateRoomParams{
		WriteRoomParams: params,
		Verified:        false,
		CreatedBy:       reqID,
	}
	r, err := repo.GormRepo.CreateRoom(p)
	return r, defaultErrorHandling(err)
}

func (repo *service) CreateVerifiedRoom(ctx context.Context, params domain.WriteRoomParams) (*domain.Room, error) {
	reqID, _ := domain.GetUserID(ctx)

	if !repo.IsPrivilege(ctx) {
		return nil, domain.ErrForbidden
	}
	p := db.CreateRoomParams{
		WriteRoomParams: params,
		Verified:        true,
		CreatedBy:       reqID,
	}
	r, err := repo.GormRepo.CreateRoom(p)
	return r, defaultErrorHandling(err)
}

func (repo *service) UpdateRoom(ctx context.Context, roomID uuid.UUID, params domain.WriteRoomParams) (*domain.Room, error) {
	reqID, _ := domain.GetUserID(ctx)

	if !repo.IsRoomAdmins(ctx, roomID) {
		return nil, domain.ErrForbidden
	}

	p := db.UpdateRoomParams{
		WriteRoomParams: params,
		CreatedBy:       reqID,
	}

	r, err := repo.GormRepo.UpdateRoom(roomID, p)
	return r, defaultErrorHandling(err)
}

func (repo *service) VerifyRoom(ctx context.Context, roomID uuid.UUID) error {
	if !repo.IsPrivilege(ctx) {
		return domain.ErrForbidden
	}
	err := repo.GormRepo.UpdateRoomVerified(roomID, true)
	return defaultErrorHandling(err)
}

func (repo *service) UnVerifyRoom(ctx context.Context, roomID uuid.UUID) error {
	if !repo.IsPrivilege(ctx) {
		return domain.ErrForbidden
	}
	err := repo.GormRepo.UpdateRoomVerified(roomID, false)
	return defaultErrorHandling(err)
}

func (repo *service) DeleteRoom(ctx context.Context, roomID uuid.UUID) error {
	if !repo.IsRoomAdmins(ctx, roomID) {
		return domain.ErrForbidden
	}
	err := repo.GormRepo.DeleteRoom(roomID)
	return defaultErrorHandling(err)
}

func (repo *service) GetRoom(ctx context.Context, roomID uuid.UUID, excludeEventID uuid.UUID) (*domain.Room, error) {
	rs, err := repo.GormRepo.GetRoom(roomID, excludeEventID)
	return rs, defaultErrorHandling(err)
}

func (repo *service) GetAllRooms(ctx context.Context, start time.Time, end time.Time, excludeEventID uuid.UUID) ([]*domain.Room, error) {
	rs, err := repo.GormRepo.GetAllRooms(start, end, excludeEventID)
	return rs, defaultErrorHandling(err)
}

func (repo *service) IsRoomAdmins(ctx context.Context, roomID uuid.UUID) bool {
	reqID, _ := domain.GetUserID(ctx)

	room, err := repo.GetRoom(ctx, roomID, uuid.Nil)
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
