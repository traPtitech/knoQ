package repository

import (
	"fmt"
	"os"
	"time"

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
}

var (
	MARIADB_HOSTNAME = os.Getenv("MARIADB_HOSTNAME")
	MARIADB_DATABASE = os.Getenv("MARIADB_DATABASE")
	MARIADB_USERNAME = os.Getenv("MARIADB_USERNAME")
	MARIADB_PASSWORD = os.Getenv("MARIADB_PASSWORD")

	DB        *gorm.DB
	logger, _ = zap.NewDevelopment()
)

// User traQユーザー情報構造体
type User struct {
	// TRAQID traQID
	TRAQID string `json:"traq_id" gorm:"type:varchar(32);primary_key"`
	// Admin 管理者かどうか
	Admin bool `gorm:"not null"`
}

// Tag Room Group Event have tags
type Tag struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Official bool   `json:"official"`
	Locked   bool   `json:"locked" gorm:"-"`
	ForRoom  bool   `json:"for_room"`
	ForGroup bool   `json:"for_group"`
	ForEvent bool   `json:"for_event"`
}

// EventTag is many to many table
type EventTag struct {
	TagID   int
	EventID int
	Locked  bool
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

// Event 予約情報
type Event struct {
	ID            int       `json:"id" gorm:"AUTO_INCREMENT"`
	Name          string    `json:"name" gorm:"type:varchar(32); not null"`
	Description   string    `json:"description" gorm:"type:varchar(1024)"`
	GroupID       int       `json:"group_id,omitempty" gorm:"not null"`
	Group         Group     `json:"group" gorm:"foreignkey:group_id"`
	RoomID        int       `json:"room_id,omitempty" gorm:"not null"`
	Room          Room      `json:"room" gorm:"foreignkey:room_id"`
	TimeStart     string    `json:"time_start" gorm:"type:TIME"`
	TimeEnd       string    `json:"time_end" gorm:"type:TIME"`
	CreatedBy     string    `json:"created_by" gorm:"type:varchar(32);"`
	AllowTogether bool      `json:"allow_together"`
	Tags          []Tag     `json:"tags" gorm:"many2many:event_tags"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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
	logger.Warn("DB error " + err.Error())
}
