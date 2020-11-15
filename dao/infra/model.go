package infra

import (
	"fmt"
	"os"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var tables = []interface{}{
	Room{},
	Event{},
}

type GormRepository struct {
	db *gorm.DB
}

func (repo *GormRepository) Setup() error {
	host := os.Getenv("MARIADB_HOSTNAME")
	user := os.Getenv("MARIADB_USERNAME")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("MARIADB_PASSWORD")
	if password == "" {
		password = "password"
	}

	var err error
	repo.db, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=true&loc=Local", user, password, host, "knoQ"),
		DefaultStringSize:         256,   // default size for string fields
		DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
	}), &gorm.Config{})
	if err != nil {
		return err
	}

	return repo.db.AutoMigrate(tables...)
}

type UserMeta struct {
	ID uuid.UUID `gorm:"type:char(36); primaryKey"`
	// Admin アプリの管理者かどうか
	Admin      bool   `gorm:"not null"`
	IsTraq     bool   `gorm:"not null"`
	Token      string `gorm:"type:varbinary(64)"`
	IcalSecret string `gorm:"not null"`
}
type UserBody struct {
	ID          uuid.UUID `gorm:"type:char(36); primaryKey"`
	Name        string    `gorm:"type:varchar(32);"`
	DisplayName string    `gorm:"type:varchar(32);"`
}

type Room struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey"`
	Place     string    `gorm:"type:varchar(32);"`
	Verified  bool
	TimeStart time.Time `gorm:"type:DATETIME; index"`
	TimeEnd   time.Time `gorm:"type:DATETIME; index"`
	Events    []Event   `gorm:"->; constraint:-"` // readOnly
	CreatedBy UserMeta  `gorm:"->"`
	gorm.Model
}

type Group struct {
	ID          uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name        string    `gorm:"type:varchar(32);not null"`
	Description string    `gorm:"type:TEXT"`
	JoinFreely  bool
	Members     []UserMeta `gorm:"->; many2many:group_members"`
	CreatedBy   UserMeta   `gorm:"->"`
	gorm.Model
}

type Tag struct {
	ID     uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	Name   string    `json:"name" gorm:"unique; type:varchar(16)"`
	Locked bool      `gorm:"-"` // for Event.Tags
	gorm.Model
}

type EventTag struct {
	EventID uuid.UUID `gorm:"type:char(36); primaryKey"`
	TagID   uuid.UUID `gorm:"type:char(36); primaryKey"`
	Locked  bool
}

type Event struct {
	ID            uuid.UUID `gorm:"type:char(36); primaryKey"`
	Name          string    `gorm:"type:varchar(32); not null"`
	Description   string    `gorm:"type:TEXT"`
	GroupID       uuid.UUID `gorm:"type:char(36); not null; index"`
	Group         Group     `gorm:"->; foreignKey:group_id;"`
	RoomID        uuid.UUID `gorm:"type:char(36); not null; "`
	Room          Room      `gorm:"->; foreignKey:RoomID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TimeStart     time.Time `gorm:"type:DATETIME; index"`
	TimeEnd       time.Time `gorm:"type:DATETIME; index"`
	CreatedBy     UserMeta  `gorm:"->"`
	AllowTogether bool
	Tags          []Tag `gorm:"many2many:event_tags;"`
	gorm.Model
}
