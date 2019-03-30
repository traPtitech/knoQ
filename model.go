package main

import "time"

// 新たにモデル(テーブル)を定義した場合はここに追加する事
var tables = []interface{}{
	User{},
	Room{},
	Group{},
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
	ID            int       `json:"id" gorm:"primary_key; AUTO_INCREMENT"`
	Place         string    `json:"place"`
	Date          string    `json:"date" gorm:"TIMESTAMP"`
	DateTimeStart string    `json:"dateTimeStart"`
	DateTimeEnd   string    `json:"dateTimeEnd"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Group グループ情報
type Group struct {
	ID        int       `json:"id" gorm:"primary_key; AUTO_INCREMENT"`
	Members   []User    `json:"members" gorm:"many2many:groups_users"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
