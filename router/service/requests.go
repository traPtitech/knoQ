package service

import (
	"time"

	"github.com/gofrs/uuid"
)

// GroupReq is group request model
type GroupReq struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	JoinFreely  bool        `json:"open"`
	Members     []uuid.UUID `json:"members"`
	Admins      []uuid.UUID `json:"admins"`
}

// RoomReq is room request model
type RoomReq struct {
	Place     string    `json:"place"`
	TimeStart time.Time `json:"timeStart"`
	TimeEnd   time.Time `json:"timeEnd"`
}

type TagReq struct {
	Name string `json:"name"`
}
type TagRelationReq struct {
	Name   string `json:"name"`
	Locked bool   `json:"locked"`
}

type EventReq struct {
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	AllowTogether bool             `json:"sharedRoom"`
	TimeStart     time.Time        `json:"timeStart"`
	TimeEnd       time.Time        `json:"timeEnd"`
	RoomID        uuid.UUID        `json:"roomId"`
	GroupID       uuid.UUID        `json:"groupId"`
	Tags          []TagRelationReq `json:"tags"`
	Admins        []uuid.UUID      `json:"admins"`
}
