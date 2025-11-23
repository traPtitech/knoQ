package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
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

type contextKey string

const (
	userIDKey contextKey = "userID"
)

// Context に UserID をセットするヘルパー
func SetUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// Context から UserID を取得するヘルパー
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
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
}
