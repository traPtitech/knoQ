package infra

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

var tables = []interface{}{
	UserMeta{},
	UserBody{},
	Group{},
	Tag{},
	Room{},
	Event{},
	EventTag{}, // Eventより下にないと、overrideされる
}

type UserMeta struct {
	ID uuid.UUID `gorm:"type:char(36); primaryKey"`
	// Admin アプリの管理者かどうか
	Admin      bool   `gorm:"not null"`
	IsTraq     bool   `gorm:"not null"`
	Token      string `gorm:"type:varbinary(64)"`
	IcalSecret string `gorm:"not null"`
}
type UserBody struct {
	ID          uuid.UUID `gorm:"type:char(36); primaryKey;"`
	Name        string    `gorm:"type:varchar(32);"`
	DisplayName string    `gorm:"type:varchar(32);"`
	UserMeta    UserMeta  `gorm:"->; foreignKey:ID; constraint:OnDelete:CASCADE;"`
}

type Room struct {
	ID             uuid.UUID `gorm:"type:char(36);primaryKey"`
	Place          string    `gorm:"type:varchar(32);"`
	Verified       bool
	TimeStart      time.Time `gorm:"type:DATETIME; index"`
	TimeEnd        time.Time `gorm:"type:DATETIME; index"`
	Events         []Event   `gorm:"->; constraint:-"` // readOnly
	CreatedByRefer uuid.UUID `gorm:"type:char(36);"`
	CreatedBy      UserMeta  `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;"`
	gorm.Model     `cvt:"->"`
}

type Group struct {
	ID             uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name           string    `gorm:"type:varchar(32);not null"`
	Description    string    `gorm:"type:TEXT"`
	JoinFreely     bool
	Members        []UserMeta `gorm:"->; many2many:group_members"`
	CreatedByRefer uuid.UUID  `gorm:"type:char(36);"`
	CreatedBy      UserMeta   `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;"`
	gorm.Model     `cvt:"->"`
}

type Tag struct {
	ID         uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name       string    `gorm:"unique; type:varchar(16)"`
	Locked     bool      `gorm:"-"` // for Event.Tags
	gorm.Model `cvt:"->"`
}

type EventTag struct {
	EventID uuid.UUID `gorm:"type:char(36); primaryKey"`
	TagID   uuid.UUID `gorm:"type:char(36); primaryKey"`
	Event   Event     `gorm:"->; foreignKey:EventID; constraint:OnDelete:CASCADE;"`
	Tag     Tag       `gorm:"->; foreignKey:TagID; constraint:OnDelete:CASCADE;"`
	Locked  bool
}

// Event is event for gorm
//go:generate go run github.com/fuji8/gotypeconverter/cmd/type-converter -s domain.WriteEventParams -d Event -o converter.go .
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
	CreatedByRefer uuid.UUID `gorm:"type:char(36);"`
	CreatedBy      UserMeta  `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;"`
	AllowTogether  bool
	Tags           []Tag `gorm:"many2many:event_tags;"`
	gorm.Model     `cvt:"->"`
}
