package domain

import (
	"time"

	"github.com/gofrs/uuid"
)

var (
	DEVELOPMENT bool
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

type Repository interface {
	EventRepository
	GroupRepository
	RoomRepository
	TagRepository
	UserRepository
}
