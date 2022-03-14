// Package v8 7db45ee2fea28db8308b72488bc18519cdf2b55f infra/db/model.go
package v8

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

var Tables = []interface{}{
	User{},
	UserBody{},
	Token{},
	Provider{},
	Group{},
	GroupMember{},
	GroupAdmin{},
	Tag{},
	Room{},
	RoomAdmin{},
	Event{},
	EventTag{}, // Eventより下にないと、overrideされる
	EventAdmin{},
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
	UserID uuid.UUID `gorm:"type:char(36); primaryKey"`

	*Oauth2Token
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

type RoomAdmin struct {
	UserID uuid.UUID `gorm:"type:char(36); primaryKey"`
	RoomID uuid.UUID `gorm:"type:char(36); primaryKey"`
	User   User      `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
	Model  `cvt:"-"`
}

// Room is
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
	Model          `cvt:"->"`
}
