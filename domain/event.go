package domain

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain/filter"
)

type ScheduleStatus int

const (
	Pending ScheduleStatus = iota + 1
	Attendance
	Absent
)

type Event struct {
	ID            uuid.UUID
	Name          string
	Description   string
	IsRoomEvent   bool
	Room          *Room
	Venue         sql.NullString
	Group         Group
	TimeStart     time.Time
	TimeEnd       time.Time
	CreatedBy     User
	Admins        []User
	Tags          []EventTag
	AllowTogether bool
	Attendees     []Attendee
	Open          bool
	Model
}

type EventTag struct {
	Tag    Tag
	Locked bool
}

type Attendee struct {
	UserID   uuid.UUID
	Schedule ScheduleStatus
}

// for repository

// WriteEventParams is used create and update
type WriteEventParams struct {
	Name          string
	Description   string
	IsRoomEvent   bool
	GroupID       uuid.UUID
	RoomID        uuid.NullUUID
	Venue         sql.NullString // option
	TimeStart     time.Time
	TimeEnd       time.Time
	Admins        []uuid.UUID
	Tags          []EventTagParams
	AllowTogether bool
	Open          bool
}

type EventTagParams struct {
	Name   string
	Locked bool
}

// EventRepository is implemented by ...
type EventRepository interface {
	CreateEvent(eventParams WriteEventParams, info *ConInfo) (*Event, error)

	UpdateEvent(eventID uuid.UUID, eventParams WriteEventParams, info *ConInfo) (*Event, error)
	AddEventTag(eventID uuid.UUID, tagName string, locked bool, info *ConInfo) error

	DeleteEvent(eventID uuid.UUID, info *ConInfo) error
	// DeleteTagInEvent delete a tag in that Event
	DeleteEventTag(eventID uuid.UUID, tagName string, info *ConInfo) error

	UpsertMeEventSchedule(eventID uuid.UUID, schedule ScheduleStatus, info *ConInfo) error

	GetEvent(eventID uuid.UUID, info *ConInfo) (*Event, error)
	GetEvents(expr filter.Expr, info *ConInfo) ([]*Event, error)
	IsEventAdmins(eventID uuid.UUID, info *ConInfo) bool

	// GetEventActivities(day int) ([]*Event, error)
}

func (e *Event) TimeConsistency() bool {
	return e.TimeStart.Before(e.TimeEnd)
}

func (e *Event) RoomTimeConsistency() bool {
	times := e.Room.CalcAvailableTime(e.AllowTogether)
	for _, t := range times {
		start := t.TimeStart
		end := t.TimeEnd
		if start.Equal(e.TimeStart) || start.Before(e.TimeStart) &&
			(end.Equal(e.TimeEnd) || end.After(e.TimeEnd)) {
			return true
		}
	}
	return false
}

func (e *Event) AdminsValidation() bool {
	return len(e.Admins) != 0
}
