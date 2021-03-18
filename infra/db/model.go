package db

import (
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

var tables = []interface{}{
	UserMeta{},
	UserBody{},
	Token{},
	Provider{},
	GroupMember{},
	GroupAdmin{},
	Group{},
	Tag{},
	Room{},
	Event{},
	EventTag{}, // Eventより下にないと、overrideされる
	EventAdmin{},
}

type Token struct {
	UserID uuid.UUID `gorm:"type:char(36); primaryKey"`
	*oauth2.Token
}

type Provider struct {
	UserID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	Issuer  string    `gorm:"not null"`
	Subject string
}

type UserMeta struct {
	ID uuid.UUID `gorm:"type:char(36); primaryKey"`
	// アプリの管理者かどうか
	Privilege  bool     `gorm:"not null"`
	IcalSecret string   `gorm:"not null"`
	Provider   Provider `gorm:"foreignKey:UserID; constraint:OnDelete:CASCADE;"`
	Token      Token    `gorm:"foreignKey:UserID; constraint:OnDelete:CASCADE;"`
}

type UserBody struct {
	ID          uuid.UUID `gorm:"type:char(36); primaryKey;"`
	Name        string    `gorm:"type:varchar(32);"`
	DisplayName string    `gorm:"type:varchar(32);"`
	Icon        string
	UserMeta    UserMeta `gorm:"->; foreignKey:ID; constraint:OnDelete:CASCADE;" cvt:"->"`
}

// Room is
//go:generate gotypeconverter -s writeRoomParams -d Room -o converter.go .
type Room struct {
	ID             uuid.UUID `gorm:"type:char(36);primaryKey"`
	Place          string    `gorm:"type:varchar(32);"`
	Verified       bool
	TimeStart      time.Time `gorm:"type:DATETIME; index"`
	TimeEnd        time.Time `gorm:"type:DATETIME; index"`
	Events         []Event   `gorm:"->; constraint:-"` // readOnly
	CreatedByRefer uuid.UUID `gorm:"type:char(36);" cvt:"CreatedBy, <-"`
	CreatedBy      UserMeta  `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;" cvt:"->"`
	gorm.Model     `cvt:"->"`
}

type GroupMember struct {
	UserID   uuid.UUID `gorm:"type:char(36); primaryKey"`
	GroupID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	UserMeta UserMeta  `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
}

type GroupAdmin struct {
	UserID   uuid.UUID `gorm:"type:char(36); primaryKey"`
	GroupID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	UserMeta UserMeta  `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
}

// Group is user group
//go:generate gotypeconverter -s writeGroupParams -d Group -o converter.go .
//go:generate gotypeconverter -s Group -d domain.Group -o converter.go .
type Group struct {
	ID             uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name           string    `gorm:"type:varchar(32);not null"`
	Description    string    `gorm:"type:TEXT"`
	JoinFreely     bool
	Members        []GroupMember
	Admins         []GroupAdmin
	CreatedByRefer uuid.UUID `gorm:"type:char(36);" cvt:"CreatedBy, <-"`
	CreatedBy      UserMeta  `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;" cvt:"->"`
	gorm.Model     `cvt:"->"`
}

type Tag struct {
	ID         uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name       string    `gorm:"unique; type:varchar(16) binary"`
	gorm.Model `cvt:"->"`
}

// EventTag is
//go:generate gotypeconverter -s domain.WriteTagRelationParams -d EventTag -o converter.go .
type EventTag struct {
	TagID   uuid.UUID `gorm:"type:char(36); primaryKey" cvt:"ID"`
	EventID uuid.UUID `gorm:"type:char(36); primaryKey"`
	Event   Event     `gorm:"->; foreignKey:EventID; constraint:OnDelete:CASCADE;"`
	Tag     Tag       `gorm:"foreignKey:TagID; constraint:OnDelete:CASCADE;" cvt:"write:Name"`
	Locked  bool
}

type EventAdmin struct {
	UserID   uuid.UUID `gorm:"type:char(36); primaryKey"`
	EventID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	UserMeta UserMeta  `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
}

// Event is event for gorm
//go:generate gotypeconverter -s writeEventParams -d Event -o converter.go .
//go:generate gotypeconverter -s Event -d domain.Event -o converter.go .
type Event struct {
	ID             uuid.UUID `gorm:"type:char(36); primaryKey"`
	Name           string    `gorm:"type:varchar(32); not null"`
	Description    string    `gorm:"type:TEXT"`
	GroupID        uuid.UUID `gorm:"type:char(36); not null; index"`
	Group          Group     `gorm:"->; foreignKey:GroupID; constraint:-"`
	RoomID         uuid.UUID `gorm:"type:char(36); not null; "`
	Room           Room      `gorm:"->; foreignKey:RoomID; constraint:OnDelete:CASCADE;"`
	TimeStart      time.Time `gorm:"type:DATETIME; index"`
	TimeEnd        time.Time `gorm:"type:DATETIME; index"`
	CreatedByRefer uuid.UUID `gorm:"type:char(36); not null" cvt:"CreatedBy, <-"`
	CreatedBy      UserMeta  `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;" cvt:"->"`
	Admins         []EventAdmin
	AllowTogether  bool
	Tags           []EventTag
	gorm.Model     `cvt:"->"`
}
