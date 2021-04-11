package db

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

func roomFullPreload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Events").Preload("Admins")
}

type WriteRoomParams struct {
	domain.WriteRoomParams

	Verified  bool
	CreatedBy uuid.UUID
}

func (repo GormRepository) CreateRoom(params WriteRoomParams) (*domain.Room, error) {
	room, err := createRoom(repo.db, params)
	if err != nil {
		return nil, err
	}
	r := ConvRoomTodomainRoom(*room)
	return &r, nil
}

func (repo GormRepository) UpdateRoom(roomID uuid.UUID, params WriteRoomParams) (*domain.Room, error) {
	room, err := updateRoom(repo.db, roomID, params)
	if err != nil {
		return nil, err
	}
	r := ConvRoomTodomainRoom(*room)
	return &r, nil
}

func (repo GormRepository) UpdateRoomVerified(roomID uuid.UUID, verified bool) error {
	return updateRoomVerified(repo.db, roomID, verified)
}

func (repo GormRepository) DeleteRoom(roomID uuid.UUID) error {
	return deleteRoom(repo.db, roomID)
}

func (repo GormRepository) GetRoom(roomID uuid.UUID) (*domain.Room, error) {
	room, err := getRoom(roomFullPreload(repo.db), roomID)
	if err != nil {
		return nil, err
	}
	r := ConvRoomTodomainRoom(*room)
	return &r, nil
}

func (repo GormRepository) GetAllRooms(start, end time.Time) ([]*domain.Room, error) {
	rooms, err := getAllRooms(roomFullPreload(repo.db), start, end)
	if err != nil {
		return nil, err
	}
	r := ConvSPRoomToSPdomainRoom(rooms)
	return r, nil
}

func createRoom(db *gorm.DB, roomParams WriteRoomParams) (*Room, error) {
	room := ConvWriteRoomParamsToRoom(roomParams)
	err := db.Create(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func updateRoom(db *gorm.DB, roomID uuid.UUID, params WriteRoomParams) (*Room, error) {
	room := ConvWriteRoomParamsToRoom(params)
	room.ID = roomID
	err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&room).Error
	return &room, err
}

func updateRoomVerified(db *gorm.DB, roomID uuid.UUID, verified bool) error {
	return db.Model(&Room{}).Where("id = ?", roomID).Update("verified", verified).Error
}

func deleteRoom(db *gorm.DB, roomID uuid.UUID) error {
	room := Room{
		ID: roomID,
	}
	err := db.Delete(&room).Error
	return err
}

func getRoom(db *gorm.DB, roomID uuid.UUID) (*Room, error) {
	room := Room{}
	err := db.Take(&room, roomID).Error
	return &room, err
}

func getAllRooms(db *gorm.DB, start, end time.Time) ([]*Room, error) {
	rooms := make([]*Room, 0)
	if !start.IsZero() {
		db = db.Where("time_start >= ?", start)
	}
	if !end.IsZero() {
		db = db.Where("time_end <= ?", end)
	}
	err := db.Order("time_start").Find(&rooms).Error
	return rooms, err
}
