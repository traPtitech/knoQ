package domain

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain/filter"
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
	Model
}

type EventTag struct {
	Tag
	Locked bool
}

// for repository

// WriteEventParams is used create and update
type WriteEventParams struct {
	Name          string
	Description   string
	GroupID       uuid.UUID
	RoomID        uuid.UUID
	TimeStart     time.Time
	TimeEnd       time.Time
	Admins        []uuid.UUID
	AllowTogether bool
	Tags          []EventTagParams
}

type EventTagParams struct {
	Name   string
	Locked bool
}

// WriteTagRelationParams is used create and update
type WriteTagRelationParams struct {
	ID     uuid.UUID
	Locked bool
}

// EventRepository is implemented by ...
type EventRepository interface {
	CreateEvent(eventParams WriteEventParams, info *ConInfo) (*Event, error)

	UpdateEvent(eventID uuid.UUID, eventParams WriteEventParams, info *ConInfo) (*Event, error)
	AddTagToEvent(eventID uuid.UUID, tagID uuid.UUID, locked bool, info *ConInfo) error

	DeleteEvent(eventID uuid.UUID, info *ConInfo) error
	// DeleteTagInEvent delete a tag in that Event
	DeleteTagInEvent(eventID uuid.UUID, tagID uuid.UUID, info *ConInfo) error

	GetEvent(eventID uuid.UUID) (*Event, error)
	GetEvents(filter.Expr) ([]*Event, error)

	GetEventActivities(day int) ([]*Event, error)
}