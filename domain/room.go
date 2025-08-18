package domain

import (
	"time"

	"github.com/gofrs/uuid"
)

type Room struct {
	ID        uuid.UUID
	Name      string
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
	// CreateUnVerifiedRoom(params WriteRoomParams, info *ConInfo) (*Room, error)
	CreateVerifiedRoom(params WriteRoomParams, info *ConInfo) (*Room, error)

	UpdateRoom(roomID uuid.UUID, params WriteRoomParams, info *ConInfo) (*Room, error)
	VerifyRoom(roomID uuid.UUID, info *ConInfo) error
	UnVerifyRoom(roomID uuid.UUID, info *ConInfo) error

	DeleteRoom(roomID uuid.UUID, info *ConInfo) error

	GetRoom(roomID uuid.UUID, excludeEventID uuid.UUID) (*Room, error)
	GetAllRooms(start time.Time, end time.Time, excludeEventID uuid.UUID) ([]*Room, error)
	IsRoomAdmins(roomID uuid.UUID, info *ConInfo) bool
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

// as: 利用可能な時間帯のリスト
// b: イベントの時間
func timeRangesSub(as []StartEndTime, b StartEndTime) (cs []StartEndTime) {
	for _, a := range as {
		cs = append(cs, timeRangeSub(a, b)...)
	}
	return
}

func timeRangeSub(a StartEndTime, b StartEndTime) []StartEndTime {
	/*
		a: ---s#####e---
		b: s##########e-
		-> -------------
	*/
	if b.TimeStart.Unix() <= a.TimeStart.Unix() && b.TimeEnd.Unix() >= a.TimeEnd.Unix() {
		return nil
	}

	/*
		a: s####e-------    a: -------s####e
		b: -------s####e    b: s####e-------
		-> s####e-------    -> -------s####e
	*/
	if a.TimeEnd.Unix() <= b.TimeStart.Unix() || b.TimeEnd.Unix() <= a.TimeStart.Unix() {
		return []StartEndTime{a}
	}

	// 期間 b が 期間 a に包含される場合
	if a.TimeStart.Unix() <= b.TimeStart.Unix() && b.TimeEnd.Unix() <= a.TimeEnd.Unix() {
		/*
			a: --s######e---
			b: --s####e-----
			-> -------s#e---
		*/
		if a.TimeStart.Unix() == b.TimeStart.Unix() {
			return []StartEndTime{
				{b.TimeEnd, a.TimeEnd},
			}
		}

		/*
			a: --s######e---
			b: ----s####e---
			-> --s#e--------
		*/
		if a.TimeEnd.Unix() == b.TimeEnd.Unix() {
			return []StartEndTime{
				{a.TimeStart, b.TimeStart},
			}
		}

		/*
			a: -s##########e
			b: ----s####e---
			-> -s##e----s##e
		*/
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

func (r *Room) TimeConsistency() bool {
	return r.TimeStart.Before((r.TimeEnd))
}

func (r *Room) AdminsValidation() bool {
	return len(r.Admins) != 0
}
