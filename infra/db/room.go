package db

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

func roomFullPreload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Events")
}

type WriteRoomParams struct {
	domain.WriteRoomParams

	Verified  bool
	CreatedBy uuid.UUID
}

func createRoom(db *gorm.DB, roomParams WriteRoomParams) (*Room, error) {
	room := ConvertWriteRoomParamsToRoom(roomParams)
	err := db.Create(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func updateRoom(db *gorm.DB, roomID uuid.UUID, params WriteRoomParams) (*Room, error) {
	room := ConvertWriteRoomParamsToRoom(params)
	room.ID = roomID
	err := db.Save(&room).Error
	return &room, err
}

func updateVerified(db *gorm.DB, roomID uuid.UUID, verified bool) error {
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
	cmd := roomFullPreload(db)
	if !start.IsZero() {
		cmd = cmd.Where("time_start >= ?", start)
	}
	if !end.IsZero() {
		cmd = cmd.Where("time_end <= ?", end)
	}
	err := cmd.Order("time_start").Find(&rooms).Error
	return rooms, err
}
