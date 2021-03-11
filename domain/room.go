package domain

import (
	"time"

	"github.com/gofrs/uuid"
)

type Room struct {
	ID    uuid.UUID
	Place string
	// Verifeid indicates if the room has been verified by privileged users.
	Verified  bool
	TimeStart time.Time
	TimeEnd   time.Time
	Events    []Event
	CreatedBy User
	Model
}

type WriteRoomParams struct {
	Place string

	// Verifeid indicates if the room has been verified by privileged users.
	TimeStart time.Time
	TimeEnd   time.Time
}

type RoomRepository interface {
	CreateRoom(roomParams WriteRoomParams, info *ConInfo) (*Room, error)
	UpdateRoom(roomID uuid.UUID, roomParams WriteRoomParams, info *ConInfo) (*Room, error)
	DeleteRoom(roomID uuid.UUID, info *ConInfo) error
	GetRoom(roomID uuid.UUID) (*Room, error)
	GetAllRooms(start *time.Time, end *time.Time) ([]*Room, error)
}
