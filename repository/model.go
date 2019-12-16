package repository

import (
	"time"
)

// Model is defalut
type Model struct {
	ID        uint64 `gorm:"primary_key"`
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
	// TRAQID traQID
	TRAQID string `json:"traq_id" gorm:"type:varchar(32);primary_key"`
	// Admin 管理者かどうか
	Admin bool `gorm:"not null"`
}

// Tag Room Group Event have tags
type Tag struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Official bool   `json:"official"`
	Locked   bool   `json:"locked" gorm:"-"`
	ForRoom  bool   `json:"for_room"`
	ForGroup bool   `json:"for_group"`
	ForEvent bool   `json:"for_event"`
}

// EventTag is many to many table
type EventTag struct {
	TagID   uint64 `gorm:"primary_key"`
	EventID uint64 `gorm:"primary_key"`
	Locked  bool
}

// GroupTag is many to many table
type GroupTag struct {
	TagID   uint64 `gorm:"primary_key"`
	GroupID uint64 `gorm:"primary_key"`
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
	Name        string `json:"name" gorm:"type:varchar(32);not null"`
	Description string `json:"description" gorm:"type:varchar(1024)"`
	Members     []User `json:"members" gorm:"many2many:group_users; save_associations:false"`
	CreatedBy   string `json:"created_by" gorm:"type:varchar(32);"`
	Tags        []Tag  `json:"tags" gorm:"many2many:group_tags; save_associations:false"`
}

// Event 予約情報
type Event struct {
	Model
	Name          string `json:"name" gorm:"type:varchar(32); not null"`
	Description   string `json:"description" gorm:"type:varchar(1024)"`
	GroupID       uint64 `json:"group_id,omitempty" gorm:"not null"`
	Group         Group  `json:"group" gorm:"foreignkey:group_id; save_associations:false"`
	RoomID        uint64 `json:"room_id,omitempty" gorm:"not null"`
	Room          Room   `json:"room" gorm:"foreignkey:room_id; save_associations:false"`
	TimeStart     string `json:"time_start" gorm:"type:TIME"`
	TimeEnd       string `json:"time_end" gorm:"type:TIME"`
	CreatedBy     string `json:"created_by" gorm:"type:varchar(32);"`
	AllowTogether bool   `json:"allow_together"`
	Tags          []Tag  `json:"tags" gorm:"many2many:event_tags; save_associations:false"`
}
