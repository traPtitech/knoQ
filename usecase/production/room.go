package production

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
)

func (repo *Repository) CreateUnVerifiedRoom(params domain.WriteRoomParams, info *domain.ConInfo) (*domain.Room, error) {
	p := db.WriteRoomParams{
		WriteRoomParams: params,
		Verified:        false,
		CreatedBy:       info.ReqUserID,
	}
	return repo.gormRepo.CreateRoom(p)
}

func (repo *Repository) CreateVerifiedRoom(params domain.WriteRoomParams, info *domain.ConInfo) (*domain.Room, error) {
	if !repo.IsPrevilege(info) {
		return nil, domain.ErrForbidden
	}
	p := db.WriteRoomParams{
		WriteRoomParams: params,
		Verified:        true,
		CreatedBy:       info.ReqUserID,
	}
	return repo.gormRepo.CreateRoom(p)
}

func (repo *Repository) UpdateRoom(roomID uuid.UUID, params domain.WriteRoomParams, info *domain.ConInfo) (*domain.Room, error) {
	if !repo.IsRoomAdmins(roomID, info) {
		return nil, domain.ErrForbidden
	}

	p := db.WriteRoomParams{
		WriteRoomParams: params,
		CreatedBy:       info.ReqUserID,
	}

	return repo.gormRepo.UpdateRoom(roomID, p)
}

func (repo *Repository) VerifyRoom(roomID uuid.UUID, info *domain.ConInfo) error {
	if !repo.IsPrevilege(info) {
		return domain.ErrForbidden
	}
	return repo.gormRepo.UpdateRoomVerified(roomID, true)
}

func (repo *Repository) UnVerifyRoom(roomID uuid.UUID, info *domain.ConInfo) error {
	if !repo.IsPrevilege(info) {
		return domain.ErrForbidden
	}
	return repo.gormRepo.UpdateRoomVerified(roomID, false)
}

func (repo *Repository) DeleteRoom(roomID uuid.UUID, info *domain.ConInfo) error {
	if !repo.IsRoomAdmins(roomID, info) {
		return domain.ErrForbidden
	}
	return repo.gormRepo.DeleteRoom(roomID)
}

func (repo *Repository) GetRoom(roomID uuid.UUID) (*domain.Room, error) {
	return repo.gormRepo.GetRoom(roomID)
}

func (repo *Repository) GetAllRooms(start time.Time, end time.Time) ([]*domain.Room, error) {
	return repo.gormRepo.GetAllRooms(start, end)
}

func (repo *Repository) IsRoomAdmins(roomID uuid.UUID, info *domain.ConInfo) bool {
	room, err := repo.GetRoom(roomID)
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
