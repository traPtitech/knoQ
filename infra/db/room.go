package db

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

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

func getRoom(db *gorm.DB, roomID uuid.UUID) (*Room, error) {
	room := Room{
		ID: roomID,
	}
	err := db.Take(&room).Error
	return &room, err
}
