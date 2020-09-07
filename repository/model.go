package repository

import (
	"time"

	"github.com/gofrs/uuid"
)

// Model is defalut
type Model struct {
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updateAt"`
	DeletedAt *time.Time `json:"-" sql:"index"`
}

// StartEndTime has start and end time
type StartEndTime struct {
	TimeStart time.Time `json:"timeStart" gorm:"type:TIME;"`
	TimeEnd   time.Time `json:"timeEnd" gorm:"type:TIME;"`
}

// UserMeta knoQユーザー情報構造体
// 当サービスの中で意味があるユーザー情報
type UserMeta struct {
	ID uuid.UUID `gorm:"type:char(36); primary_key"`
	// Admin アプリの管理者かどうか
	Admin bool   `gorm:"not null"`
	Token string `gorm:"type:varbinary(64)"`
}

// UserBody ユーザー情報
// 現在、DBには存在しない
type UserBody struct {
	ID          uuid.UUID
	Name        string
	DisplayName string
}

// Tag Room Group Event have tags
type Tag struct {
	ID     uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	Name   string    `json:"name" gorm:"unique; type:varchar(16)"`
	Locked bool      `gorm:"-"`
	Model
}

// EventTag is many to many table
type EventTag struct {
	TagID   uuid.UUID `gorm:"type:char(36); primary_key"`
	EventID uuid.UUID `gorm:"type:char(36); primary_key"`
	Locked  bool
}

// GroupUsers is many to many table
type GroupUsers struct {
	GroupID uuid.UUID `gorm:"type:char(36); primary_key;not null"`
	UserID  uuid.UUID `gorm:"type:char(36); primary_key;not null"`
}

// Room 部屋情報
type Room struct {
	ID        uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	Place     string    `json:"place" gorm:"type:varchar(32);"`
	Public    bool
	TimeStart time.Time `json:"timeStart" gorm:"type:DATETIME; index"`
	TimeEnd   time.Time `json:"timeEnd" gorm:"type:DATETIME; index"`
	Events    []Event   `gorm:"foreignkey:RoomID"`
	CreatedBy uuid.UUID `gorm:"type:char(36)"`
	Model
}

// Group グループ情報
// Group is not user JSON
type Group struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key"`
	Name        string    `gorm:"type:varchar(32);not null"`
	Description string    `gorm:"type:TEXT"`
	JoinFreely  bool
	Members     []GroupUsers `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:GroupID"`
	CreatedBy   uuid.UUID    `gorm:"type:char(36);"`
	Model
}

// Event 予約情報
type Event struct {
	ID          uuid.UUID `json:"eventId" gorm:"type:char(36);primary_key"`
	Name        string    `json:"name" gorm:"type:varchar(32); not null"`
	Description string    `json:"description" gorm:"type:TEXT"`
	GroupID     uuid.UUID `json:"groupId" gorm:"type:char(36);not null; index"`
	//Group         Group     `json:"-" gorm:"foreignkey:group_id; save_associations:false"`
	RoomID        uuid.UUID `json:"roomId" gorm:"type:char(36);not null"`
	Room          Room      `json:"-" gorm:"foreignkey:room_id; save_associations:false"`
	TimeStart     time.Time `json:"timeStart" gorm:"type:DATETIME; index"`
	TimeEnd       time.Time `json:"timeEnd" gorm:"type:DATETIME; index"`
	CreatedBy     uuid.UUID `json:"createdBy" gorm:"type:char(36);"`
	AllowTogether bool      `json:"sharedRoom"`
	Tags          []Tag     `json:"tags" gorm:"many2many:event_tags; association_autoupdate:false;association_autocreate:false"`
	Model
}
