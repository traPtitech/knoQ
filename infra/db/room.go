package db

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

func roomFullPreload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("events")
}

type writeRoomParams struct {
	domain.WriteRoomParams

	Verified  bool
	CreatedBy uuid.UUID
}

func createRoom(db *gorm.DB, roomParams writeRoomParams) (*Room, error) {
	room := ConvertwriteRoomParamsToRoom(roomParams)
	err := db.Create(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func updateRoom(db *gorm.DB, roomID uuid.UUID, params writeRoomParams) (*Room, error) {
	room := ConvertwriteRoomParamsToRoom(params)
	room.ID = roomID
	err := db.Save(&room).Error
	return &room, err
}

func deleteRoom(db *gorm.DB, roomID uuid.UUID) error {
	room := Room{
		ID: roomID,
	}
	err := db.Delete(&room).Error
	return err
}

func getRoom(db *gorm.DB, roomID uuid.UUID) (*Room, error) {
	room := Room{
		ID: roomID,
	}
	cmd := roomFullPreload(db)
	err := cmd.Take(&room).Error
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
