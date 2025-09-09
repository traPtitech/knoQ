package db

import (
	"time"

	gomysql "github.com/go-sql-driver/mysql"
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filters"
	"github.com/traPtitech/knoQ/migration"
	"golang.org/x/oauth2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormRepository

type GormRepository interface {
	Setup(host, user, password, database, port, key, logLevel string, loc *time.Location) error
	CreateEvent(params WriteEventParams) (*Event, error)
	UpdateEvent(eventID uuid.UUID, params WriteEventParams) (*Event, error)
	AddEventTag(eventID uuid.UUID, params domain.EventTagParams) error
	DeleteEvent(eventID uuid.UUID) error
	DeleteEventTag(eventID uuid.UUID, tagName string, deleteLocked bool) error
	UpsertEventSchedule(eventID, userID uuid.UUID, scheduleStatus domain.ScheduleStatus) error
	GetEvent(eventID uuid.UUID) (*Event, error)
	GetAllEvents(expr filters.Expr) ([]*Event, error)
	CreateGroup(params WriteGroupParams) (*Group, error)
	UpdateGroup(groupID uuid.UUID, params WriteGroupParams) (*Group, error)
	AddMemberToGroup(groupID, userID uuid.UUID) error
	DeleteGroup(groupID uuid.UUID) error
	DeleteMemberOfGroup(groupID, userID uuid.UUID) error
	GetGroup(groupID uuid.UUID) (*Group, error)
	GetAllGroups() ([]*Group, error)
	GetBelongGroupIDs(userID uuid.UUID) ([]uuid.UUID, error)
	GetAdminGroupIDs(userID uuid.UUID) ([]uuid.UUID, error)
	CreateRoom(params CreateRoomParams) (*domain.Room, error)
	UpdateRoom(roomID uuid.UUID, params UpdateRoomParams) (*domain.Room, error)
	UpdateRoomVerified(roomID uuid.UUID, verified bool) error
	DeleteRoom(roomID uuid.UUID) error
	GetRoom(roomID uuid.UUID, excludeEventID uuid.UUID) (*domain.Room, error)
	GetAllRooms(start, end time.Time, excludeEventID uuid.UUID) ([]*domain.Room, error)
	CreateOrGetTag(name string) (*domain.Tag, error)
	GetTag(tagID uuid.UUID) (*domain.Tag, error)
	GetAllTags() ([]*domain.Tag, error)
	GetToken(userID uuid.UUID) (*oauth2.Token, error)
	SaveUser(user User) (*User, error)
	UpdateiCalSecret(userID uuid.UUID, secret string) error
	GetUser(userID uuid.UUID) (*User, error)
	GetAllUsers(onlyActive bool) ([]*User, error)
	SyncUsers(users []*User) error
	GrantPrivilege(userID uuid.UUID) error
}

// structのポインタを返す
func NewGormRepository() GormRepository {
	gormrepo := gormRepository{}
	return &gormrepo
}

type gormRepository struct {
	db *gorm.DB
}

var tokenKey []byte

func (repo *gormRepository) Setup(host, user, password, database, port, key, logLevel string, loc *time.Location) error {
	loglevel := func() logger.LogLevel {
		switch logLevel {
		case "slient":
			return logger.Silent
		case "error":
			return logger.Error
		case "info":
			return logger.Info
		case "warn":
			return logger.Warn
		default:
			return logger.Silent
		}
	}()
	if len(key) != 32 {
		panic("token key is not 32 words")
	}
	tokenKey = []byte(key)

	var err error
	repo.db, err = gorm.Open(
		mysql.New(mysql.Config{
			DSNConfig: &gomysql.Config{
				User:                 user,
				Passwd:               password,
				Net:                  "tcp",
				Addr:                 host + ":" + port,
				DBName:               database,
				Loc:                  loc,
				AllowNativePasswords: true,
				ParseTime:            true,
			},
			DefaultStringSize:         256,   // default size for string fields
			DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
			DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
			DontSupportRenameColumn:   false, // `change` when rename column, rename column not supported before MySQL 8, MariaDB
			SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
		}),
		&gorm.Config{
			Logger: logger.Default.LogMode(loglevel),
		})
	if err != nil {
		return err
	}

	return migration.Migrate(repo.db, tables)
}
