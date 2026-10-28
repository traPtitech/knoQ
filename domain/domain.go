package domain

import (
	"context"
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

type TransactionManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service interface {
	DraftEventService
	EventService
	GroupService
	RoomService
	TagService
	UserService
}

type Repository interface {
	DraftEventRepository
	EventRepository
	GroupRepository
	RoomRepository
	TagRepository
	UserRepository
}
