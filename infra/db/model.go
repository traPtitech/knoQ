package db

import (
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

var tables = []interface{}{
	User{},
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

type Model struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
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

type User struct {
	ID uuid.UUID `gorm:"type:char(36); primaryKey"`
	// アプリの管理者かどうか
	Privilege  bool `gorm:"<-:create; not null"` // Do not update
	State      int
	IcalSecret string   `gorm:"not null"`
	Provider   Provider `gorm:"foreignKey:UserID; constraint:OnDelete:CASCADE;"`
	Token      Token    `gorm:"foreignKey:UserID; constraint:OnDelete:CASCADE;"`
}

type UserBody struct {
	ID          uuid.UUID `gorm:"type:char(36); primaryKey;"`
	Name        string    `gorm:"type:varchar(32);"`
	DisplayName string    `gorm:"type:varchar(32);"`
	Icon        string
	User        User `gorm:"->; foreignKey:ID; constraint:OnDelete:CASCADE;" cvt:"->"`
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
	CreatedBy      User      `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;" cvt:"->"`
	Model          `cvt:"->"`
}

//go:generate gotypeconverter -s []*GroupMember -d []uuid.UUID -o converter.go -structTag cvt0 .
type GroupMember struct {
	UserID  uuid.UUID `gorm:"type:char(36); primaryKey" cvt0:"<-"`
	GroupID uuid.UUID `gorm:"type:char(36); primaryKey"`
	User    User      `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
}

type GroupAdmin struct {
	UserID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	GroupID uuid.UUID `gorm:"type:char(36); primaryKey"`
	User    User      `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
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
	CreatedBy      User      `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;" cvt:"->"`
	Model          `cvt:"->"`
}

type Tag struct {
	ID    uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name  string    `gorm:"unique; type:varchar(16) binary"`
	Model `cvt:"->"`
}

// EventTag is
type EventTag struct {
	TagID   uuid.UUID `gorm:"type:char(36); primaryKey" cvt:"ID"`
	EventID uuid.UUID `gorm:"type:char(36); primaryKey"`
	Event   Event     `gorm:"->; foreignKey:EventID; constraint:OnDelete:CASCADE;"`
	Tag     Tag       `gorm:"foreignKey:TagID; constraint:OnDelete:CASCADE;" cvt:"write:Name"`
	Locked  bool
	Model   `cvt:"->"`
}

type EventAdmin struct {
	UserID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	EventID uuid.UUID `gorm:"type:char(36); primaryKey"`
	User    User      `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
	Model   `cvt:"-"`
}

// Event is event for gorm
//go:generate gotypeconverter -s WriteEventParams -d Event -o converter.go .
//go:generate gotypeconverter -s Event -d domain.Event -o converter.go .
type Event struct {
	ID             uuid.UUID `gorm:"type:char(36); primaryKey"`
	Name           string    `gorm:"type:varchar(32); not null"`
	Description    string    `gorm:"type:TEXT"`
	GroupID        uuid.UUID `gorm:"type:char(36); not null; index"`
	Group          Group     `gorm:"->; foreignKey:GroupID; constraint:-"`
	RoomID         uuid.UUID `gorm:"type:char(36); not null; index"`
	Room           Room      `gorm:"->; foreignKey:RoomID; constraint:OnDelete:CASCADE;"`
	TimeStart      time.Time `gorm:"type:DATETIME; index"`
	TimeEnd        time.Time `gorm:"type:DATETIME; index"`
	CreatedByRefer uuid.UUID `gorm:"type:char(36); not null" cvt:"CreatedBy, <-"`
	CreatedBy      User      `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;" cvt:"->"`
	Admins         []EventAdmin
	AllowTogether  bool
	Tags           []EventTag
	Model          `cvt:"->"`
}
