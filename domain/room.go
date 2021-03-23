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

// StartEndTime has start and end time
type StartEndTime struct {
	TimeStart time.Time
	TimeEnd   time.Time
}

// CalcAvailableTime calclate available time
// allowTogether = true 併用化の時間帯
// allowTogether = false 誰も取っていない時間帯
func (r *Room) CalcAvailableTime(allowTogether bool) []StartEndTime {
	availabletime := []StartEndTime{
		{
			TimeStart: r.TimeStart,
			TimeEnd:   r.TimeEnd,
		},
	}
	for _, e := range r.Events {
		if allowTogether && e.AllowTogether {
			continue
		}
		availabletime = timeRangesSub(availabletime, StartEndTime{e.TimeStart, e.TimeEnd})
	}
	return availabletime
}

func timeRangesSub(as []StartEndTime, b StartEndTime) (cs []StartEndTime) {
	for _, a := range as {
		cs = append(cs, timeRangeSub(a, b)...)
	}
	return
}

func timeRangeSub(a StartEndTime, b StartEndTime) []StartEndTime {
	/*
		a: s####e-------
		b: -------s####e
		-> s####e
	*/
	if a.TimeStart.Unix() >= b.TimeEnd.Unix() || a.TimeEnd.Unix() <= b.TimeEnd.Unix() {
		return []StartEndTime{a}
	}

	/*
		a: ---s#####e---
		b: s##########e-
		-> -------------
	*/
	if b.TimeStart.Unix() <= a.TimeStart.Unix() && b.TimeEnd.Unix() >= a.TimeEnd.Unix() {
		return nil
	}

	/*
		a: s###########e
		b: ----s####e---
		-> s###e----s##e
	*/
	if a.TimeStart.Unix() < b.TimeStart.Unix() && b.TimeEnd.Unix() < a.TimeEnd.Unix() {
		return []StartEndTime{
			{a.TimeStart, b.TimeStart},
			{b.TimeEnd, a.TimeEnd},
		}
	}

	/*
		a: s#####e------
		b: ----s######e-
		-> s###e--------
	*/
	if a.TimeStart.Unix() < b.TimeStart.Unix() && a.TimeEnd.Unix() < b.TimeEnd.Unix() {
		return []StartEndTime{
			{a.TimeStart, b.TimeStart},
		}
	}

	/*
		a: -----s######e
		b: --s#####e----
		-> --------s###e
	*/
	if b.TimeStart.Unix() < a.TimeStart.Unix() && b.TimeEnd.Unix() < a.TimeEnd.Unix() {
		return []StartEndTime{
			{b.TimeEnd, a.TimeEnd},
		}
	}
	return nil
}
