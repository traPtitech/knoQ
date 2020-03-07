// Package repository is
package repository

import (
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"

	// mysql
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// 新たにモデル(テーブル)を定義した場合はここに追加する事
var tables = []interface{}{
	User{},
	Room{},
	Group{},
	Event{},
	Tag{},
	EventTag{},
	UserSession{},
}

type Repository interface {
	GroupRepository
	RoomRepository
}

// GormRepository implements Repository interface
type GormRepository struct {
	DB *gorm.DB
}

// APIRepository implements only GroupRepository interface
type APIRepository struct {
	url string
}

var (
	MARIADB_HOSTNAME = os.Getenv("MARIADB_HOSTNAME")
	MARIADB_DATABASE = os.Getenv("MARIADB_DATABASE")
	MARIADB_USERNAME = os.Getenv("MARIADB_USERNAME")
	MARIADB_PASSWORD = os.Getenv("MARIADB_PASSWORD")

	DB        *gorm.DB
	logger, _ = zap.NewDevelopment()
)

// CRUD is create, read, update, delete
// all need ID
type CRUD interface {
	Create() error
	Read() error
	// Update update omitempty
	Update() error
	Delete() error
}

// SetupDatabase set up DB and crate tables
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
	DB, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", MARIADB_USERNAME, MARIADB_PASSWORD, MARIADB_HOSTNAME, MARIADB_DATABASE))
	if err != nil {
		return DB, err
	}
	if err := initDB(); err != nil {
		return DB, err
	}
	return DB, nil
}

// initDB データベースのスキーマを更新
func initDB() error {
	// テーブルが無ければ作成
	if err := DB.AutoMigrate(tables...).Error; err != nil {
		return err
	}
	return nil
}

func dbErrorLog(err error) {
	if gorm.IsRecordNotFoundError(err) {
		return
	}
	me, ok := err.(*mysql.MySQLError)
	if !ok {
		logger.Warn("DB error " + err.Error())
		return
	}
	if me.Number == 1062 {
		return
	}

	logger.Warn("DB error " + err.Error())
}
