package presentation

import (
	"time"

	"github.com/gofrs/uuid"
)

//go:generate go run github.com/fuji8/gotypeconverter/cmd/type-converter -s RoomReq -d domain.WriteRoomParams -o converter.go .
type RoomReq struct {
	Place     string    `json:"place"`
	TimeStart time.Time `json:"timeStart"`
	TimeEnd   time.Time `json:"timeEnd"`
}

type StartEndTime struct {
	TimeStart time.Time `json:"timeStart"`
	TimeEnd   time.Time `json:"timeEnd"`
}

//go:generate go run github.com/fuji8/gotypeconverter/cmd/type-converter -s domain.Room -d RoomRes -o converter.go .
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
