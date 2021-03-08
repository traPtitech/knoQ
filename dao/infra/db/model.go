package db

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

var tables = []interface{}{
	UserMeta{},
	UserBody{},
	GroupMember{},
	GroupAdmins{},
	Group{},
	Tag{},
	Room{},
	Event{},
	EventTag{}, // Eventより下にないと、overrideされる
	EventAdmin{},
}

type UserMeta struct {
	ID uuid.UUID `gorm:"type:char(36); primaryKey"`
	// Admin アプリの管理者かどうか
	Privilege  bool   `gorm:"not null"`
	IsTraq     bool   `gorm:"not null"`
	Token      string `gorm:"type:varbinary(64)"`
	IcalSecret string `gorm:"not null"`
}
type UserBody struct {
	ID          uuid.UUID `gorm:"type:char(36); primaryKey;"`
	Name        string    `gorm:"type:varchar(32);"`
	DisplayName string    `gorm:"type:varchar(32);"`
	UserMeta    UserMeta  `gorm:"->; foreignKey:ID; constraint:OnDelete:CASCADE;" cvt:"->"`
}

// Room is
//go:generate go run github.com/fuji8/gotypeconverter/cmd/type-converter -s writeRoomParams -d Room -o converter.go .
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

type GroupAdmins struct {
	UserID   uuid.UUID `gorm:"type:char(36); primaryKey"`
	GroupID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	UserMeta UserMeta  `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
}

// Group is user group
//go:generate go run github.com/fuji8/gotypeconverter/cmd/type-converter -s writeGroupParams -d Group -o converter.go .
//go:generate go run github.com/fuji8/gotypeconverter/cmd/type-converter -s Group -d domain.Group -o converter.go .
type Group struct {
	ID             uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name           string    `gorm:"type:varchar(32);not null"`
	Description    string    `gorm:"type:TEXT"`
	JoinFreely     bool
	Members        []GroupMember
	Admins         []GroupAdmins
	CreatedByRefer uuid.UUID `gorm:"type:char(36);" cvt:"CreatedBy, <-"`
	CreatedBy      UserMeta  `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;" cvt:"->"`
	gorm.Model     `cvt:"->"`
}

type Tag struct {
	ID         uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name       string    `gorm:"unique; type:varchar(16)"`
	Locked     bool      `gorm:"-"` // for Event.Tags
	gorm.Model `cvt:"->"`
}

type EventTag struct {
	TagID   uuid.UUID `gorm:"type:char(36); primaryKey"`
	EventID uuid.UUID `gorm:"type:char(36); primaryKey"`
	Event   Event     `gorm:"->; foreignKey:EventID; constraint:OnDelete:CASCADE;"`
	Tag     Tag       `gorm:"foreignKey:TagID; constraint:OnDelete:CASCADE;" cvt:"Name"`
	Locked  bool
}

type EventAdmin struct {
	UserID   uuid.UUID `gorm:"type:char(36); primaryKey"`
	EventID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	UserMeta UserMeta  `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
}

// Event is event for gorm
//go:generate go run github.com/fuji8/gotypeconverter/cmd/type-converter -s writeEventParams -d Event -o converter.go .
//go:generate go run github.com/fuji8/gotypeconverter/cmd/type-converter -s Event -d domain.Event -o converter.go .
type Event struct {
	ID             uuid.UUID `gorm:"type:char(36); primaryKey"`
	Name           string    `gorm:"type:varchar(32); not null"`
	Description    string    `gorm:"type:TEXT"`
	GroupID        uuid.UUID `gorm:"type:char(36); not null; index"`
	Group          Group     `gorm:"->; foreignKey:group_id; constraint:-"`
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
