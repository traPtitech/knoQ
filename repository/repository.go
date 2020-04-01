// Package repository is
package repository

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jwt"

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
	GroupUsers{},
}

type Repository interface {
	UserRepository
	GroupRepository
	RoomRepository
	EventRepository
	TagRepository
}

// GormRepository implements Repository interface
type GormRepository struct {
	DB *gorm.DB
}

type TraQVersion int64

const (
	V1 TraQVersion = iota
	V3
)

var traQEndPoints = [2]string{
	"https://q.trap.jp/api/1.0",
	"https://q.trap.jp/api/v3",
}

type TraQRepository struct {
	Version TraQVersion
	Token   string
}

type GoogleAPIRepository struct {
	Config     *jwt.Config
	Client     *http.Client
	CalendarID string
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

func (repo *GoogleAPIRepository) Setup() {
	repo.Client = repo.Config.Client(oauth2.NoContext)
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
	if err := initDB(DB); err != nil {
		return DB, err
	}
	return DB, nil
}

// initDB データベースのスキーマを更新
func initDB(db *gorm.DB) error {
	// gormのエラーの上書き
	gorm.ErrRecordNotFound = ErrNotFound

	// テーブルが無ければ作成
	if err := db.AutoMigrate(tables...).Error; err != nil {
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
