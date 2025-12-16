package domain

import (
	"time"
)

var (
	VERSION     = "UNKNOWN"
	REVISION    = "UNKNOWN"
	DEVELOPMENT bool
)

type Model struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Service interface {
	EventService
	GroupService
	RoomService
	TagService
	UserService
}

type Repository interface {
	EventRepository
	GroupRepository
	RoomRepository
	TagRepository
	UserRepository
}
