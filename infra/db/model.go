package db

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

var tables = []interface{}{
	User{},
	Group{},
	GroupMember{},
	GroupAdmin{},
	Tag{},
	Room{},
	RoomAdmin{},
	Event{},
	EventTag{}, // Eventより下にないと、overrideされる
	EventAdmin{},
	EventAttendee{},
}

type Model struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Oauth2Token is oauth2.Token
type Oauth2Token struct {
	// AccessToken is the token that authorizes and authenticates
	// the requests.
	AccessToken string `gorm:"type:varbinary(64)"`

	// TokenType is the type of token.
	// The Type method returns either this or "Bearer", the default.
	TokenType string

	// RefreshToken is a token that's used by the application
	// (as opposed to the user) to refresh the access token
	// if it expires.
	RefreshToken string

	// Expiry is the optional expiration time of the access token.
	//
	// If zero, TokenSource implementations will reuse the same
	// token forever and RefreshToken or equivalent
	// mechanisms for that TokenSource will not be used.
	Expiry time.Time
}
type Token struct {
	*Oauth2Token
}

type Provider struct {
	Issuer  string `gorm:"not null"`
	Subject string
}

type User struct {
	ID uuid.UUID `gorm:"type:char(36); primaryKey"`
	// アプリの管理者かどうか
	Privilege  bool `gorm:"not null"`
	State      int
	IcalSecret string `gorm:"not null"`
	Provider
	*Oauth2Token
}

// type UserBody struct {
// 	ID          uuid.UUID `gorm:"type:char(36); primaryKey;"`
// 	Name        string    `gorm:"type:varchar(32);"`
// 	DisplayName string    `gorm:"type:varchar(32);"`
// 	Icon        string
// 	User        User `gorm:"->; foreignKey:ID; constraint:OnDelete:CASCADE;" cvt:"->"`
// }

type RoomAdmin struct {
	UserID uuid.UUID `gorm:"type:char(36); primaryKey"`
	RoomID uuid.UUID `gorm:"type:char(36); primaryKey"`
	User   User      `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
	Model  `cvt:"-"`
}

//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s Room -d domain.Room -o converter.go .
//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s []*Room -d []*domain.Room -o converter.go .
type Room struct {
	ID             uuid.UUID `gorm:"type:char(36);primaryKey"`
	Place          string    `gorm:"type:varchar(32);"`
	Verified       bool
	TimeStart      time.Time `gorm:"type:DATETIME; index"`
	TimeEnd        time.Time `gorm:"type:DATETIME; index"`
	Events         []Event   `gorm:"->; constraint:-"` // readOnly
	Admins         []RoomAdmin
	CreatedByRefer uuid.UUID `gorm:"type:char(36);" cvt:"CreatedBy, <-"`
	CreatedBy      User      `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;" cvt:"->"`
	Model          `cvt:"->"`
}

//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s []*GroupMember -d []uuid.UUID -o converter.go -structTag cvt0 .
type GroupMember struct {
	UserID  uuid.UUID `gorm:"type:char(36); primaryKey" cvt0:"<-"`
	GroupID uuid.UUID `gorm:"type:char(36); primaryKey"`
	User    User      `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
	Model   `cvt:"-"`
}

type GroupAdmin struct {
	UserID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	GroupID uuid.UUID `gorm:"type:char(36); primaryKey"`
	User    User      `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
	Model   `cvt:"-"`
}

// Group is user group
//
//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s Group -d domain.Group -o converter.go .
//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s []*Group -d []*domain.Group -o converter.go .
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

//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s Tag -d domain.Tag -o converter.go .
//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s []*Tag -d []*domain.Tag -o converter.go .
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

//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s []EventAdmin -d []RoomAdmin -o converter.go .
type EventAdmin struct {
	UserID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	EventID uuid.UUID `gorm:"type:char(36); primaryKey"`
	User    User      `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
	Model   `cvt:"-"`
}

type EventAttendee struct {
	UserID   uuid.UUID `gorm:"type:char(36); primaryKey"`
	EventID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	User     User      `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
	Schedule int
}

// Event is event for gorm
//
//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s Event -d domain.Event -o converter.go .
//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s []*Event -d []*domain.Event -o converter.go .
type Event struct {
	ID             uuid.UUID `gorm:"type:char(36); primaryKey"`
	Name           string    `gorm:"type:varchar(32); not null"`
	Description    string    `gorm:"type:TEXT"`
	GroupID        uuid.UUID `gorm:"type:char(36); not null; index"`
	Group          Group     `gorm:"->; foreignKey:GroupID; constraint:-"`
	RoomID         uuid.UUID `gorm:"type:char(36); not null; index"`
	Room           Room      `gorm:"foreignKey:RoomID; constraint:OnDelete:CASCADE;" cvt:"write:Place"`
	TimeStart      time.Time `gorm:"type:DATETIME; index"`
	TimeEnd        time.Time `gorm:"type:DATETIME; index"`
	CreatedByRefer uuid.UUID `gorm:"type:char(36); not null" cvt:"CreatedBy, <-"`
	CreatedBy      User      `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;" cvt:"->"`
	Admins         []EventAdmin
	AllowTogether  bool
	Tags           []EventTag
	Open           bool
	Attendees      []EventAttendee
	Model          `cvt:"->"`
}
