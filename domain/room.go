package domain

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/utils"
)

type Room struct {
	ID    uuid.UUID
	Place string
	// Verified indicates if the room has been verified by privileged users.
	Verified  bool
	TimeStart time.Time
	TimeEnd   time.Time
	Events    []Event
	Admins    []User
	CreatedBy User
	Model
}

type WriteRoomParams struct {
	Place string

	// Verifeid indicates if the room has been verified by privileged users.
	TimeStart time.Time
	TimeEnd   time.Time

	Admins []uuid.UUID
}

type RoomRepository interface {
	CreateUnVerifiedRoom(params WriteRoomParams, info *ConInfo) (*Room, error)
	CreateVerifiedRoom(params WriteRoomParams, info *ConInfo) (*Room, error)

	UpdateRoom(roomID uuid.UUID, params WriteRoomParams, info *ConInfo) (*Room, error)
	VerifyRoom(roomID uuid.UUID, info *ConInfo) error
	UnVerifyRoom(roomID uuid.UUID, info *ConInfo) error

	DeleteRoom(roomID uuid.UUID, info *ConInfo) error

	GetRoom(roomID uuid.UUID, excludeEventID uuid.UUID) (*Room, error)
	GetAllRooms(start time.Time, end time.Time, excludeEventID uuid.UUID) ([]*Room, error)
	IsRoomAdmins(roomID uuid.UUID, info *ConInfo) bool
}

func (r *Room) CalcAvailableTime(allowTogether bool) []utils.TimeSpan {
	available := []utils.TimeSpan{
		{
			Start: r.TimeStart,
			End:   r.TimeEnd,
		},
	}

	for _, e := range r.Events {
		if allowTogether && e.AllowTogether {
			continue
		}
		eventTime := utils.TimeSpan{
			Start: e.TimeStart,
			End:   e.TimeEnd,
		}
		available = subtractMultipleTimeSpans(available, eventTime)
	}

	return available
}

// 複数の時間帯から 1 つの時間帯を引く
func subtractMultipleTimeSpans(bases []utils.TimeSpan, sub utils.TimeSpan) []utils.TimeSpan {
	var result []utils.TimeSpan
	for _, base := range bases {
		result = append(result, utils.SubtractTimeSpan(base, sub)...)
	}
	return result
}

func (r *Room) TimeConsistency() bool {
	return r.TimeStart.Before((r.TimeEnd))
}

func (r *Room) AdminsValidation() bool {
	return len(r.Admins) != 0
}
