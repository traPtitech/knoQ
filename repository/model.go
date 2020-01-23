package repository

import (
	"time"

	"github.com/gofrs/uuid"
)

// Model is defalut
type Model struct {
	ID        uuid.UUID  `json:"id" gorm:"type:char(36);primary_key"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updateAt"`
	DeletedAt *time.Time `json:"-" sql:"index"`
}

// StartEndTime has start and end time
type StartEndTime struct {
	TimeStart string `json:"timeStart" gorm:"type:TIME;"`
	TimeEnd   string `json:"timeEnd" gorm:"type:TIME;"`
}

// User traQユーザー情報構造体
type User struct {
	// ID traQID
	ID uuid.UUID `json:"id" gorm:"type:char(36); primary_key" traq:"userId"`
	// Admin 管理者かどうか
	Admin bool `json:"admin" gorm:"not null"`
	// tmp
	Auth string `json:"-" gorm:"-"`
}

// UserSession has user session
type UserSession struct {
	Token         string    `gorm:"primary_key; type:char(32);"`
	UserID        uuid.UUID `gorm:"type:char(36);"`
	Authorization string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time `sql:"index"`
}

// Tag Room Group Event have tags
type Tag struct {
	Model
	Name     string `json:"name" gorm:"unique; type:varchar(16)"`
	Official bool   `json:"official"`
	Locked   bool   `json:"locked" gorm:"-"`
}

// EventTag is many to many table
type EventTag struct {
	TagID   uuid.UUID `gorm:"type:char(36); primary_key"`
	EventID uuid.UUID `gorm:"type:char(36); primary_key"`
	Locked  bool
}

// Room 部屋情報
type Room struct {
	Model
	Place         string         `json:"place" gorm:"type:varchar(16);unique_index:idx_room_unique"`
	Date          string         `json:"date" gorm:"type:DATE; unique_index:idx_room_unique"`
	TimeStart     string         `json:"timeStart" gorm:"type:TIME; unique_index:idx_room_unique"`
	TimeEnd       string         `json:"timeEnd" gorm:"type:TIME; unique_index:idx_room_unique"`
	AvailableTime []StartEndTime `json:"availableTime" gorm:"-"`
}

// Group グループ情報
// Group is not user JSON
type Group struct {
	Model
	Name        string    `json:"name" gorm:"type:varchar(32);not null"`
	Description string    `json:"description" gorm:"type:varchar(1024)"`
	ImageID     string    `json:"imageId"`
	JoinFreely  bool      `json:"open"`
	Members     []User    `json:"members" gorm:"many2many:group_users; save_associations:false"`
	IsTraQGroup bool      `json:"isTraQGroup" gorm:"-"`
	CreatedBy   uuid.UUID `json:"createdBy" gorm:"type:char(36);"`
}

// Event 予約情報
type Event struct {
	Model
	Name          string    `json:"name" gorm:"type:varchar(32); not null"`
	Description   string    `json:"description" gorm:"type:varchar(1024)"`
	GroupID       uuid.UUID `json:"groupId" gorm:"type:char(36);not null"`
	Group         Group     `json:"-" gorm:"foreignkey:group_id; save_associations:false"`
	RoomID        uuid.UUID `json:"rooId" gorm:"type:char(36);not null"`
	Room          Room      `json:"-" gorm:"foreignkey:room_id; save_associations:false"`
	TimeStart     string    `json:"timeStart" gorm:"type:TIME"`
	TimeEnd       string    `json:"timeEnd" gorm:"type:TIME"`
	CreatedBy     uuid.UUID `json:"createdBy" gorm:"type:char(36);"`
	AllowTogether bool      `json:"sharedRoom"`
	Tags          []Tag     `json:"tags" gorm:"many2many:event_tags; save_associations:false"`
}
