package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type DraftEventInviteeScheduleStatus int
const (
	DraftEventPending = iota + 1
	DraftEventAvailable
	DraftEventUnavailable
)

type DraftEvent struct {
	ID                 uuid.UUID
	Name               string
	Description        string
	Group              Group
	Place              string
	Admins             []User
	DeadLine           time.Time
	Tags               []DraftEventTag
	Invitees           []User
	Open               bool
	TimeWindowStates []TimeWindowInviteeStatus
	CreatedBy          User
	Model
}

type DraftEventInviteeStatus struct {
	TimeWindowID uuid.UUID
	TimeStart time.Time
	TimeEnd time.Time
	InviteesStatus [](User,)
}

type DraftEventTag struct {
	Tag    Tag
	Locked bool
}

type WriteDraftEventParams struct {
	Name        string
	Description string
	Group       Group
	Place       string
	Admins      []User
	DeadLine    time.Time
	Tags        []DraftEventTag
	Invitees    []User
	TimeWindows []DraftEventTimeWindow
	Open        bool
}

type DraftEventTimeWindow struct {
	TimeStart      time.Time
	TimeEnd        time.Time
}

type DraftEventService interface {
	CraeteDraftEvent(ctx context.Context, requesterID uuid.UUID, draftEventParams WriteDraftEventParams) (*DraftEvent, error)
	UpdateDraftEvent(ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID, draftEventParams WriteDraftEventParams) (*DraftEvent, error)
	DeleteDraftEvent(ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID) error

	FinalizeEvent(ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID) (*Event, error)

	AddDraftEventTag(ctx context.Context, reqesterID uuid.UUID, draftEventID uuid.UUID, tagName string, locked bool) error
	DeleteDraftEventTag(ctx context.Context, reqID uuid.UUID, draftEventID uuid.UUID) error

	GetDraftEvent(ctx context.Context, draftEventID uuid.UUID) (*DraftEvent, error)
	GetDraftEvents(ctx context.Context, draftEventID uuid.UUID) ([]*DraftEvent, error)

	UpsertMeDraftEventAvailability (ctx context.Context, requesterID uuid.UUID, draftEventID uuid.UUID, schedulde [])

	IsDraftEventAdmin(ctx context.Context, reqesterID uuid.UUID, draftEventID uuid.UUID) bool
}

type UpsertDraftEventArgs struct {
	WriteDraftEventParams
	CreatedBy uuid.UUID
}


type DraftEventRepogitory interface{
	CreateDraftEvent(ctx context.Context, args UpsertDraftEventArgs) (*DraftEvent,error)

	UpdateDraftEvent(ctx context.Context,draftEventID uuid.UUID, args UpsertDraftEventArgs) (*DraftEvent,error)

	DeleteDraftEvent()

}