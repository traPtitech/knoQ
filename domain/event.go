package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"github.com/traPtitech/knoQ/domain/filters"
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
	Room          Room
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

// WriteEventParams is used create and update
type WriteEventParams struct {
	Name          string
	Description   string
	GroupID       uuid.UUID
	RoomID        uuid.UUID
	Place         string // option
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
type EventService interface {
	CreateEvent(ctx context.Context, eventParams WriteEventParams) (*Event, error)

	UpdateEvent(ctx context.Context, eventID uuid.UUID, eventParams WriteEventParams) (*Event, error)
	AddEventTag(ctx context.Context, eventID uuid.UUID, tagName string, locked bool) error

	DeleteEvent(ctx context.Context, eventID uuid.UUID) error
	// DeleteTagInEvent delete a tag in that Event
	DeleteEventTag(ctx context.Context, eventID uuid.UUID, tagName string) error

	UpsertMeEventSchedule(ctx context.Context, eventID uuid.UUID, schedule ScheduleStatus) error

	GetEvent(ctx context.Context, eventID uuid.UUID) (*Event, error)
	GetEvents(ctx context.Context, expr filters.Expr) ([]*Event, error)
	IsEventAdmins(ctx context.Context, eventID uuid.UUID) bool

	GetEventsWithGroup(ctx context.Context, expr filters.Expr) ([]*Event, error)

	// GetEventActivities(day int) ([]*Event, error)
}

type CreateEventArgs struct {
	ID            uuid.UUID
	CreatedBy     uuid.UUID
	Name          string
	Description   string
	GroupID       uuid.UUID
	RoomID        uuid.UUID
	TimeStart     time.Time
	TimeEnd       time.Time
	Admins        []uuid.UUID
	Tags          []EventTagParams
	AllowTogether bool
	Open          bool
}

type UpdateEventArgs struct {
	CreatedBy     uuid.UUID
	Name          string
	Description   string
	GroupID       uuid.UUID
	RoomID        uuid.UUID
	TimeStart     time.Time
	TimeEnd       time.Time
	Admins        []uuid.UUID
	Tags          []EventTagParams
	AllowTogether bool
	Open          bool
}

type EventRepository interface {
	// 進捗部屋でない部屋は事前に作成しておく必要がある
	CreateEvent(args CreateEventArgs) (*Event, error)

	// 進捗部屋でない部屋は事前に作成しておく必要がある
	UpdateEvent(eventID uuid.UUID, args UpdateEventArgs) (*Event, error)

	AddEventTag(eventID uuid.UUID, params EventTagParams) error

	DeleteEvent(eventID uuid.UUID) error

	DeleteEventTag(eventID uuid.UUID, tagName string, deleteLocked bool) error

	UpsertEventSchedule(eventID, userID uuid.UUID, scheduleStatus ScheduleStatus) error

	GetEvent(eventID uuid.UUID) (*Event, error)

	GetAllEvents(expr filters.Expr) ([]*Event, error)
}
