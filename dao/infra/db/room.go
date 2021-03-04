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

// BeforeCreate is hook
func (r *Room) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}

func createRoom(db *gorm.DB, roomParams writeRoomParams) (*Room, error) {
	room := ConvertwriteRoomParamsToRoom(roomParams)
	err := db.Create(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}
