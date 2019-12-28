package repository

import (
	"github.com/gofrs/uuid"
	"time"
)

// Model is defalut
type Model struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

// StartEndTime has start and end time
type StartEndTime struct {
	TimeStart string `json:"time_start" gorm:"type:TIME;"`
	TimeEnd   string `json:"time_end" gorm:"type:TIME;"`
}

// User traQユーザー情報構造体
type User struct {
	// ID traQID
	ID uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	// Admin 管理者かどうか
	Admin bool `json:"admin" gorm:"not null"`
}

// Tag Room Group Event have tags
type Tag struct {
	Model
	Name     string `json:"name"`
	Official bool   `json:"official"`
	Locked   bool   `json:"locked" gorm:"-"`
	ForRoom  bool   `json:"for_room"`
	ForGroup bool   `json:"for_group"`
	ForEvent bool   `json:"for_event"`
}

// EventTag is many to many table
type EventTag struct {
	TagID   uuid.UUID `gorm:"type:char(36);primary_key"`
	EventID uuid.UUID `gorm:"type:char(36);primary_key"`
	Locked  bool
}

// Room 部屋情報
type Room struct {
	Model
	Place         string         `json:"place" gorm:"type:varchar(16);unique_index:idx_room_unique"`
	Date          string         `json:"date" gorm:"type:DATE; unique_index:idx_room_unique"`
	TimeStart     string         `json:"time_start" gorm:"type:TIME; unique_index:idx_room_unique"`
	TimeEnd       string         `json:"time_end" gorm:"type:TIME; unique_index:idx_room_unique"`
	AvailableTime []StartEndTime `json:"available_time" gorm:"-"`
}

// Group グループ情報
type Group struct {
	Model
	Name        string    `json:"name" gorm:"type:varchar(32);not null"`
	Description string    `json:"description" gorm:"type:varchar(1024)"`
	ImageID     string    `json:"image_id"`
	JoinFreely  bool      `json:"join_freely"`
	Members     []User    `json:"members" gorm:"many2many:group_users; save_associations:false"`
	IsTraQGroup bool      `json:"is_traQ_group" gorm:"-"`
	CreatedBy   uuid.UUID `json:"created_by" gorm:"type:char(36);"`
}

// Event 予約情報
type Event struct {
	Model
	Name          string    `json:"name" gorm:"type:varchar(32); not null"`
	Description   string    `json:"description" gorm:"type:varchar(1024)"`
	GroupID       uuid.UUID `json:"group_id" gorm:"type:char(36);not null"`
	Group         Group     `json:"group" gorm:"foreignkey:group_id; save_associations:false"`
	RoomID        uuid.UUID `json:"room_id" gorm:"type:char(36);not null"`
	Room          Room      `json:"room" gorm:"foreignkey:room_id; save_associations:false"`
	TimeStart     string    `json:"time_start" gorm:"type:TIME"`
	TimeEnd       string    `json:"time_end" gorm:"type:TIME"`
	CreatedBy     uuid.UUID `json:"created_by" gorm:"type:char(36);"`
	AllowTogether bool      `json:"allow_together"`
	Tags          []Tag     `json:"tags" gorm:"many2many:event_tags; save_associations:false"`
}
