package presentation

import (
	"time"

	"github.com/traPtitech/knoQ/domain"

	"github.com/gofrs/uuid"
)

type DraftEventReqCreate struct {
	Name           string                       `json:"name"`
	Description    string                       `json:"description"`
	Deadline       time.Time                    `json:"deadline"`
	Admins         []uuid.UUID                  `json:"admins"`
	Invitees       []uuid.UUID                  `json:"invitees"`
	Open           bool                         `json:"open"`
	Tags           []DraftEventTagReq           `json:"tags"`
	CandidateSlots []DraftEventCandidateSlotReq `json:"candidateSlots"`
}

type DraftEventCandidateSlotReq struct {
	Date      string `json:"date"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type DraftEventTagReq struct {
	Name   string `json:"name"`
	Locked bool   `json:"locked"`
}

// PUTリクエスト
type DraftEventReqUpdate struct {
	Name           string                       `json:"name"`
	Description    string                       `json:"description"`
	Deadline       time.Time                    `json:"deadline"`
	Admins         []uuid.UUID                  `json:"admins"`
	Invitees       []uuid.UUID                  `json:"invitees"`
	Open           bool                         `json:"open"`
	Tags           []DraftEventTagReq           `json:"tags"`
	AdditionalCandidateSlots []DraftEventCandidateSlotReq `json:"candidateSlots"`
}

type DraftEventRes struct {
	ID               uuid.UUID   `json:"draftEventId"`
	Name             string      `json:"name"`
	Deadline         time.Time   `json:"deadline"`
	DraftEventStatus string      `json:"status"`
	RespondedCount   int         `json:"respondedCount"`
	TotalINvitees    int         `json:"totalInvitees"`
	Admins           []uuid.UUID `json:"admins"`
	Open             bool        `json:"open"`
	Model
}

// GET への返答
type DraftEventResDetail struct {
	ID                       uuid.UUID                    `json:"draftEventId"`
	Name                     string                       `json:"name"`
	Description              string                       `json:"description"`
	Deadline                 time.Time                    `json:"deadline"`
	DraftEventStatus         string                       `json:"status"`
	RespondedCount           int                          `json:"respondedCount"`
	TotalInvitees            int                          `json:"totalInvitees"`
	Admins                   []uuid.UUID                  `json:"admins"`
	Open                     bool                         `json:"open"`
	DraftEventCandidateSlots []DraftEventCandidateSlotRes `json:"candidateSlots"`

	Model
}

type DraftEventCandidateSlotRes struct {
	ID        uuid.UUID `json:"slotId"`
	TimeStart time.Time `json:"timeStart"`
	TimeEnd   time.Time `json:"timeEnd"`
}

type DraftEventSchedulingResilts struct {
	ID      uuid.UUID    `json:"draftEventId"`
	Results []SlotResult `json:"results"`
}

type SlotResult struct {
	ID               uuid.UUID                     `json:"slotId"`
	AvailableCount   int                           `json:"availableCount"`
	AvailableUsers   []uuid.UUID                   `json:"availableUsers"`
	AvailabilityRate float32                       `json:"availabilityRate"`
	Respondents      []DraftEventRespondentSummary `json:"respondents"`
	NonRespondents   []uuid.UUID                   `json:"nonRespondents"`
}

type DraftEventRespondentSummary struct {
	UserID      uuid.UUID `json:"userId"`
	RespondedAt time.Time `json:"respondedAt"`
	Comment     string    `json:"comment"`
}

func ConvdomainDraftEventToDraftEventRes(src domain.DraftEvent) (dst DraftEventRes) {

	return
}

func ConvdomainDraftEventToDraftEventResDetail(src domain.DraftEvent) (dst DraftEventResDetail) {

	return
}
