package main

import "time"

// 新たにモデル(テーブル)を定義した場合はここに追加する事
var tables = []interface{}{
	User{},
	Room{},
	Group{},
	Reservation{},
}

// User traQユーザー情報構造体
type User struct {
	// TRAQID traQID
	TRAQID string `json:"traq_id" gorm:"type:varchar(32);primary_key"`
	// Admin 管理者かどうか
	Admin bool `gorm:"not null"`
}

// Room 部屋情報
type Room struct {
	ID        int       `json:"id" gorm:"primary_key; AUTO_INCREMENT"`
	Place     string    `json:"place" gorm:"type:varchar(16);unique_index:idx_room_unique"`
	Date      string    `json:"date" gorm:"type:DATE; unique_index:idx_room_unique"`
	TimeStart string    `json:"time_start" gorm:"type:TIME; unique_index:idx_room_unique"`
	TimeEnd   string    `json:"time_end" gorm:"type:TIME; unique_index:idx_room_unique"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Group グループ情報
type Group struct {
	ID             int       `json:"id" gorm:"primary_key; AUTO_INCREMENT"`
	Name           string    `json:"name" gorm:"type:varchar(32);unique;not null"`
	Description    string    `json:"description" gorm:"type:varchar(1024)"`
	Members        []User    `json:"members" gorm:"many2many:group_users"`
	CreatedBy      User      `json:"created_by" gorm:"foreignkey:CreatedByRefer; not null"`
	CreatedByRefer string    `json:"created_by_refer" gorm:"type:varchar(32);"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Reservation 予約情報
type Reservation struct {
	ID             int       `json:"id" gorm:"AUTO_INCREMENT"`
	Name           string    `json:"name" gorm:"type:varchar(32); not null"`
	Description    string    `json:"description" gorm:"type:varchar(1024)"`
	GroupID        int       `json:"group_id" gorm:"not null"`
	Group          Group     `json:"group" gorm:"foreignkey:group_id"`
	RoomID         int       `json:"room_id" gorm:"not null"`
	Room           Room      `json:"room" gorm:"foreignkey:room_id"`
	Date           string    `json:"date" gorm:"type:DATE; index:date"`
	TimeStart      string    `json:"time_start" gorm:"type:TIME"`
	TimeEnd        string    `json:"time_end" gorm:"type:TIME"`
	CreatedBy      User      `json:"created_by" gorm:"foreignkey:CreatedByRefer; not null"`
	CreatedByRefer string    `json:"created_by_refer" gorm:"type:varchar(32);"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
