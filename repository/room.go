package repository

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
)

func (repo *repository) CreateUnVerifiedRoom(params domain.WriteRoomParams, info *domain.ConInfo) (*domain.Room, error) {
	p := db.CreateRoomParams{
		WriteRoomParams: params,
		Verified:        false,
		CreatedBy:       info.ReqUserID,
	}
	r, err := repo.GormRepo.CreateRoom(p)
	return r, defaultErrorHandling(err)
}

func (repo *repository) CreateVerifiedRoom(params domain.WriteRoomParams, info *domain.ConInfo) (*domain.Room, error) {
	if !repo.IsPrivilege(info) {
		return nil, domain.ErrForbidden
	}
	p := db.CreateRoomParams{
		WriteRoomParams: params,
		Verified:        true,
		CreatedBy:       info.ReqUserID,
	}
	r, err := repo.GormRepo.CreateRoom(p)
	return r, defaultErrorHandling(err)
}

func (repo *repository) UpdateRoom(roomID uuid.UUID, params domain.WriteRoomParams, info *domain.ConInfo) (*domain.Room, error) {
	if !repo.IsRoomAdmins(roomID, info) {
		return nil, domain.ErrForbidden
	}

	p := db.UpdateRoomParams{
		WriteRoomParams: params,
		CreatedBy:       info.ReqUserID,
	}

	r, err := repo.GormRepo.UpdateRoom(roomID, p)
	return r, defaultErrorHandling(err)
}

func (repo *repository) VerifyRoom(roomID uuid.UUID, info *domain.ConInfo) error {
	if !repo.IsPrivilege(info) {
		return domain.ErrForbidden
	}
	err := repo.GormRepo.UpdateRoomVerified(roomID, true)
	return defaultErrorHandling(err)
}

func (repo *repository) UnVerifyRoom(roomID uuid.UUID, info *domain.ConInfo) error {
	if !repo.IsPrivilege(info) {
		return domain.ErrForbidden
	}
	err := repo.GormRepo.UpdateRoomVerified(roomID, false)
	return defaultErrorHandling(err)
}

func (repo *repository) DeleteRoom(roomID uuid.UUID, info *domain.ConInfo) error {
	if !repo.IsRoomAdmins(roomID, info) {
		return domain.ErrForbidden
	}
	err := repo.GormRepo.DeleteRoom(roomID)
	return defaultErrorHandling(err)
}

func (repo *repository) GetRoom(roomID uuid.UUID, excludeEventID uuid.UUID) (*domain.Room, error) {
	rs, err := repo.GormRepo.GetRoom(roomID, excludeEventID)
	return rs, defaultErrorHandling(err)
}

func (repo *repository) GetAllRooms(start time.Time, end time.Time, excludeEventID uuid.UUID) ([]*domain.Room, error) {
	rs, err := repo.GormRepo.GetAllRooms(start, end, excludeEventID)
	return rs, defaultErrorHandling(err)
}

func (repo *repository) IsRoomAdmins(roomID uuid.UUID, info *domain.ConInfo) bool {
	room, err := repo.GetRoom(roomID, uuid.Nil)
	if err != nil {
		return false
	}
	for _, admin := range room.Admins {
		if info.ReqUserID == admin.ID {
			return true
		}
	}
	return false
}
