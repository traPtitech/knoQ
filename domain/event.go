package domain

import (
	"time"

	"github.com/gofrs/uuid"
)

type Event struct {
	ID          uuid.UUID
	Name        string
	Description string
	Room        Room
	Group       Group
	TimeStart   time.Time
	TimeEnd     time.Time
	CreatedBy   User
	Tags        []struct {
		Tag
		Locked bool
	}
	AllowTogether bool
}
