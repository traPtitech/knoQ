package domain

import (
	"time"

	"github.com/gofrs/uuid"
)

type Model struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// ConInfo is Connection infomation
type ConInfo struct {
	// Token     string
	ReqUserID uuid.UUID
}

func (c *ConInfo) GetUserID() uuid.UUID {
	if c == nil {
		return uuid.Nil
	}
	return c.ReqUserID
}

type Repository interface {
	EventRepository
	GroupRepository
	RoomRepository
	TagRepository
	UserRepository
}
