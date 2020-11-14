package domain

import (
	"time"

	"github.com/gofrs/uuid"
)

type Model struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

type Event struct {
	ID            uuid.UUID
	Name          string
	Description   string
	Room          Room
	Group         Group
	TimeStart     time.Time
	TimeEnd       time.Time
	CreatedBy     User
	Tags          []EventTag
	AllowTogether bool
	Model
}

type EventTag struct {
	Tag
	Locked bool
}
