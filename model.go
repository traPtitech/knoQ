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
	Admin bool
}

// Room 部屋情報
type Room struct {
	ID        int       `json:"id" gorm:"primary_key; AUTO_INCREMENT"`
	Place     string    `json:"place"`
	Date      string    `json:"date" gorm:"type:DATE"`
	TimeStart string    `json:"time_start"`
	TimeEnd   string    `json:"time_end"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Group グループ情報
type Group struct {
	ID        int       `json:"id" gorm:"primary_key; AUTO_INCREMENT"`
	Name      string    `json:"name" gorm:"unique"`
	Members   []User    `json:"members" gorm:"many2many:groups_users"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Reservation 予約情報
type Reservation struct {
	ID        int       `json:"id" gorm:"AUTO_INCREMENT"`
	GroupID   int       `json:"group_id"`
	RoomID    int       `json:"room_id"`
	Date      string    `json:"date" gorm:"type:DATE"`
	TimeStart string    `json:"time_start"`
	TimeEnd   string    `json:"time_end"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
