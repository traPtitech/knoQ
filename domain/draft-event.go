package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain/filters"
)

type DraftEventEditability int

const (
	EventEditable DraftEventEditability = iota
	EventUneditable
)

type TimeWindowUserStatus int

const (
	TimeWindowAvailable TimeWindowUserStatus = iota
	TimeWindowUnavailable
)

type DraftEvent struct {
	ID               uuid.UUID
	Name             string
	Description      string
	DeadLine         time.Time
	Editability      DraftEventEditability
	Admins           []User
	Tags             []EventTag
	Invitees         []User
	Open             bool
	TimeWindowStates []TimeWindowInviteeStatus
	CreatedBy        User
	UserComments     []UserComment
	Model
}

type UserComment struct {
	UserID      uuid.UUID
	CommentedAt time.Time
	Comment     string
}

type TimeWindowInviteeStatus struct {
	TimeWindowID  uuid.UUID
	TimeStart     time.Time
	TimeEnd       time.Time
	UserResponses []TimeWindowUserResponse
}

type TimeWindowUserResponse struct {
	UserID                 uuid.UUID
	TimeWindowAvailability TimeWindowUserStatus
}

type WriteDraftEventParams struct {
	Name        string
	Description string
	Open        bool
	DeadLine    time.Time
	Admins      []User
	Invitees    []User
	Tags        []EventTag
	TimeWindows []DraftEventTimeWindow
}

type DraftEventTimeWindow struct {
	TimeStart time.Time
	TimeEnd   time.Time
}

type DraftEventService interface {
	CraeteDraftEvent(ctx context.Context, requesterID uuid.UUID, draftEventParams WriteDraftEventParams) (*DraftEvent, error)
	UpdateDraftEvent(ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID, draftEventParams WriteDraftEventParams) (*DraftEvent, error)
	DeleteDraftEvent(ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID) error

	AddDraftEventTag(ctx context.Context, reqesterID uuid.UUID, draftEventID uuid.UUID, tagName string, locked bool) error
	DeleteDraftEventTag(ctx context.Context, reqID uuid.UUID, draftEventID uuid.UUID) error

	GetDraftEvent(ctx context.Context, draftEventID uuid.UUID) (*DraftEvent, error)
	GetDraftEvents(ctx context.Context, reqesterID uuid.UUID, expr filters.Expr) ([]*DraftEvent, error)

	UpsertMeDraftEventAvailability(ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID, schedulde []WriteTimeWindowAvailability) error

	IsDraftEventAdmin(ctx context.Context, reqesterID uuid.UUID, draftEventID uuid.UUID) bool
}

type WriteTimeWindowAvailability struct {
	TimeWindowID uuid.UUID
	Status       ScheduleStatus
}

type UpsertDraftEventArgs struct {
	WriteDraftEventParams
	CreatedBy uuid.UUID
}

type DraftEventRepogitory interface {
	CreateDraftEvent(ctx context.Context, args UpsertDraftEventArgs) (*DraftEvent, error)

	UpdateDraftEvent(ctx context.Context, draftEventID uuid.UUID, args UpsertDraftEventArgs) (*DraftEvent, error)

	DeleteDraftEvent(ctx context.Context, draftEventID uuid.UUID) error

	AddEventTag(ctx context.Context, draftEventID uuid.UUID, params EventTagParams) error

	DeleteEventTag(ctx context.Context, draftEventID uuid.UUID, tagName string, deleteLocked bool) error

	UpsertDraftEventSchedule(ctx context.Context, drraftEventID uuid.UUID, userID uuid.UUID, ScheduleStatus []WriteTimeWindowAvailability) error

	GetDraftEvent(ctx context.Context, draftEventID uuid.UUID) (*DraftEvent, error)

	GetAllDraftEvents(ctx context.Context, expr filters.Expr) ([]*DraftEvent, error)
}
