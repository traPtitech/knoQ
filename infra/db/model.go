package db

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

var tables = []interface{}{
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
	EventAttendee{},
	EventRoom{},
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

//go:generate gotypeconverter -s []EventAdmin -d []RoomAdmin -o converter.go .
type RoomAdmin struct {
	UserID uuid.UUID `gorm:"type:char(36); primaryKey"`
	RoomID uuid.UUID `gorm:"type:char(36); primaryKey"`
	User   User      `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
	Model  `cvt:"-"`
}

// Room is
//go:generate gotypeconverter -s CreateRoomParams -d Room -o converter.go .
//go:generate gotypeconverter -s UpdateRoomParams -d Room -o converter.go .
//go:generate gotypeconverter -s Room -d domain.Room -o converter.go .
//go:generate gotypeconverter -s []*Room -d []*domain.Room -o converter.go .
type Room struct {
	ID       uuid.UUID `gorm:"type:char(36);primaryKey"`
	Place    string    `gorm:"type:varchar(32);"`
	Verified bool
	// AllowTogether bool
	TimeStart time.Time `gorm:"type:DATETIME; index"`
	TimeEnd   time.Time `gorm:"type:DATETIME; index"`
	// TODO:
	// Events         []Event `gorm:"->; foreignKey:ID"` // readOnly
	// Events         []Event
	// 相互からアクセス可能だから名前変えた方がいいかも
	Events         []EventRoom
	Admins         []RoomAdmin
	CreatedByRefer uuid.UUID `gorm:"type:char(36);" cvt:"CreatedBy, <-"`
	CreatedBy      User      `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;" cvt:"->"`
	Model          `cvt:"->"`
}

//go:generate gotypeconverter -s []*GroupMember -d []uuid.UUID -o converter.go -structTag cvt0 .
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
//go:generate gotypeconverter -s WriteGroupParams -d Group -o converter.go .
//go:generate gotypeconverter -s Group -d domain.Group -o converter.go .
//go:generate gotypeconverter -s []*Group -d []*domain.Group -o converter.go .
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

//go:generate gotypeconverter -s Tag -d domain.Tag -o converter.go .
//go:generate gotypeconverter -s []*Tag -d []*domain.Tag -o converter.go .
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

type EventAttendee struct {
	UserID   uuid.UUID `gorm:"type:char(36); primaryKey"`
	EventID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	User     User      `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
	Schedule int
}

// TODO
type EventRoom struct {
	// 実際のカラムは構造体じゃないやつのみ
	RoomID        uuid.UUID `gorm:"type:char(36); primaryKey"`
	EventID       uuid.UUID `gorm:"type:char(36); primaryKey"`
	AllowTogether bool
	// このレコードに書いてあるEventIDに対応するEventが消されたらこのレコードが消される
	// 無限にネストしないように->になってる？
	Event Event `gorm:"->; foreignKey:EventID; constraint:OnDelete:CASCADE;"`
	Room  Room  `gorm:"foreignKey:RoomID; constraint:OnDelete:CASCADE;"`
	Model
}

// Event is event for gorm
//go:generate gotypeconverter -s WriteEventParams -d Event -o converter.go .
//go:generate gotypeconverter -s Event -d domain.Event -o converter.go .
//go:generate gotypeconverter -s []*Event -d []*domain.Event -o converter.go .
type Event struct {
	ID             uuid.UUID `gorm:"type:char(36); primaryKey"`
	Name           string    `gorm:"type:varchar(32); not null"`
	Description    string    `gorm:"type:TEXT"`
	GroupID        uuid.UUID `gorm:"type:char(36); not null; index"`
	Group          Group     `gorm:"->; foreignKey:GroupID; constraint:-"`
	TimeStart      time.Time `gorm:"type:DATETIME; index"`
	TimeEnd        time.Time `gorm:"type:DATETIME; index"`
	CreatedByRefer uuid.UUID `gorm:"type:char(36); not null" cvt:"CreatedBy, <-"`
	CreatedBy      User      `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;" cvt:"->"`
	// TODO:
	// 相互からアクセス可能だから名前変えた方がいいかも
	Rooms []EventRoom
	// Rooms      []Room
	Admins    []EventAdmin
	Tags      []EventTag
	Open      bool
	Attendees []EventAttendee
	Model     `cvt:"->"`
}
