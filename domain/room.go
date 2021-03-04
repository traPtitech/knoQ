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
}

type WriteRoomParams struct {
	Place string

	// Verifeid indicates if the room has been verified by privileged users.
	Verified  bool
	TimeStart time.Time
	TimeEnd   time.Time
	CreatedBy uuid.UUID
}

type RoomRepository interface {
	CreateRoom(roomParams WriteRoomParams) (*Room, error)
	UpdateRoom(roomID uuid.UUID, roomParams WriteRoomParams) (*Room, error)
	DeleteRoom(roomID uuid.UUID, deletePublic bool) error
	GetRoom(roomID uuid.UUID) (*Room, error)
	GetAllRooms(start *time.Time, end *time.Time) ([]*Room, error)
}
