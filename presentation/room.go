package presentation

import (
	"time"

	"github.com/gofrs/uuid"
)

//go:generate gotypeconverter -s RoomReq -d domain.WriteRoomParams -o converter.go .
type RoomReq struct {
	Place     string      `json:"place"`
	TimeStart time.Time   `json:"timeStart"`
	TimeEnd   time.Time   `json:"timeEnd"`
	Admins    []uuid.UUID `json:"admins"`
}

type StartEndTime struct {
	TimeStart time.Time `json:"timeStart"`
	TimeEnd   time.Time `json:"timeEnd"`
}

//go:generate gotypeconverter -s domain.Room -d RoomRes -o converter.go .
//go:generate gotypeconverter -s []*domain.Room -d []*RoomRes -o converter.go .
type RoomRes struct {
	ID uuid.UUID `json:"roomId"`
	// Verifeid indicates if the room has been verified by privileged users.
	Verified bool `json:"verified"`
	RoomReq
	FreeTimes   []StartEndTime `json:"freeTimes" cvt:"-"`
	SharedTimes []StartEndTime `json:"sharedTimes" cvt:"-"`
	CreatedBy   UserRes        `json:"createdBy"`
	Model
}
