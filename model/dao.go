package model

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	MARIADB_HOSTNAME = os.Getenv("MARIADB_HOSTNAME")
	MARIADB_DATABASE = os.Getenv("MARIADB_DATABASE")
	MARIADB_USERNAME = os.Getenv("MARIADB_USERNAME")
	MARIADB_PASSWORD = os.Getenv("MARIADB_PASSWORD")

	db *gorm.DB
)

// SetupDatabase set up db and crate tables
func SetupDatabase() (*gorm.DB, error) {
	var err error
	//tmp
	if MARIADB_HOSTNAME == "" {
		MARIADB_HOSTNAME = ""
	}
	if MARIADB_DATABASE == "" {
		MARIADB_DATABASE = "room"
	}
	if MARIADB_USERNAME == "" {
		MARIADB_USERNAME = "root"
	}

	if MARIADB_PASSWORD == "" {
		MARIADB_PASSWORD = "password"
	}

	// データベース接続
	db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", MARIADB_USERNAME, MARIADB_PASSWORD, MARIADB_HOSTNAME, MARIADB_DATABASE))
	if err != nil {
		return db, err
	}
	if err := initDB(); err != nil {
		return db, err
	}
	return db, nil
}

// initDB データベースのスキーマを更新
func initDB() error {
	// テーブルが無ければ作成
	if err := db.AutoMigrate(tables...).Error; err != nil {
		return err
	}
	return nil
}
