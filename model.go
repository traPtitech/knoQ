package main

// 新たにモデル(テーブル)を定義した場合はここに追加する事
var tables = []interface{}{
	User{},
}

// User traQユーザー情報構造体
type User struct {
	// TRAQID traQID
	TRAQID string `gorm:"type:varchar(32);primary_key"`
	// Admin 管理者かどうか
	Admin bool
}
