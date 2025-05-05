package migration

import (
	"database/sql"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID uuid.UUID `gorm:"type:char(36);primaryKey"`
}

type v16Group struct {
	ID             uuid.UUID     `gorm:"type:char(36);primaryKey"`
	Name           string        `gorm:"type:varchar(32);not null"`
	Description    string        `gorm:"type:TEXT"`
	IsTraqGroup    bool          `gorm:"not null"`
	JoinFreely     sql.NullBool  `gorm:""`
	TraqID         uuid.NullUUID `gorm:""`
	Members        []*User       `gorm:"many2many:group_member;"` // 結合テーブル名を明示
	Admins         []*User       `gorm:"many2many:group_admin;"`  // 結合テーブル名を明示
	CreatedByRefer uuid.NullUUID `gorm:"type:char(36);"`
	CreatedBy      *User         `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;"`
}

type v16Room struct {
	ID     uuid.UUID `gorm:"type:char(36);primaryKey"`
	Admins []*User   `gorm:"many2many:room_admin;"` // 結合テーブル名を明示
}

func (*v16Room) TableName() string {
	return "rooms"
}

type v16EventTag struct {
	EventID uuid.UUID `gorm:"type:char(36);primaryKey"`
}

func (*v16EventTag) TableName() string {
	return "event_tags"
}

type v16Event struct {
	ID             uuid.UUID `gorm:"type:char(36); primaryKey"`
	Name           string    `gorm:"type:varchar(32); not null"`
	Description    string    `gorm:"type:TEXT"`
	GroupID        uuid.UUID `gorm:"type:char(36); not null; index"`
	Group          v16Group  `gorm:"->; foreignKey:GroupID; constraint:-"`
	RoomID         uuid.UUID `gorm:"type:char(36); not null; index"`
	Room           v16Room   `gorm:"foreignKey:RoomID; constraint:OnDelete:CASCADE;" cvt:"write:Place"`
	TimeStart      time.Time `gorm:"type:DATETIME; index"`
	TimeEnd        time.Time `gorm:"type:DATETIME; index"`
	CreatedByRefer uuid.UUID `gorm:"type:char(36); not null" cvt:"CreatedBy, <-"`
	CreatedBy      User      `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;" cvt:"->"`
	Admins         []*User   `gorm:"many2many:event_admin"`
	AllowTogether  bool
	Tags           []*v16EventTag `gorm:"references:EventID; constraint:OnDelete:CASCADE;"`
	Open           bool
	Attendees      []v16EventAttendee
}

type v16EventAttendee struct {
	UserID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	EventID uuid.UUID `gorm:"type:char(36); primaryKey"`
}

func (*v16EventAttendee) TableName() string {
	return "event_attendee"
}

func (*v16Event) TableName() string {
	return "events"
}

func (*v16Group) TableName() string {
	return "groups"
}

type v16GroupOld struct {
	ID             uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name           string    `gorm:"type:varchar(32);not null"`
	Description    string    `gorm:"type:TEXT"`
	IsTraqGroup    bool      `gorm:"not null;default:false"`
	JoinFreely     sql.NullBool
	TraqID         uuid.NullUUID `gorm:"default:null;uniqueIndex"`
	CreatedByRefer uuid.NullUUID `gorm:"type:char(36);"`
}

func (*v16GroupOld) TableName() string {
	return "groups" // テーブル名は変更されないことを明示
}

type v16GroupMember struct {
	UserID  uuid.UUID `gorm:"type:char(36);primaryKey"`
	GroupID uuid.UUID `gorm:"type:char(36);primaryKey"`
}

func (*v16GroupMember) TableName() string {
	return "group_members" // テーブル名は変更されないことを明示
}

type v16GroupAdmin struct {
	UserID  uuid.UUID `gorm:"type:char(36);primaryKey"`
	GroupID uuid.UUID `gorm:"type:char(36);primaryKey"`
}

func (*v16GroupAdmin) TableName() string {
	return "group_admins" // テーブル名は変更されないことを明示
}

// v16 は many2many リレーションシップへの変更を処理するマイグレーションです
func v16() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "16",

		Migrate: func(tx *gorm.DB) error {
			if tx.Migrator().HasTable("room_admins") {
				if err := tx.Migrator().RenameTable("room_admins", "room_admin"); err != nil {
					return err
				}
			}

			if tx.Migrator().HasTable("event_admins") {
				if err := tx.Migrator().RenameTable("event_admins", "event_admin"); err != nil {
					return err
				}
			}

			if tx.Migrator().HasTable("group_members") {
				if err := tx.Migrator().RenameTable("group_members", "group_member"); err != nil {
					return err
				}
			}
			if tx.Migrator().HasTable("group_admins") {
				if err := tx.Migrator().RenameTable("group_admins", "group_admin"); err != nil {
					return err
				}
			}
			return tx.AutoMigrate(&v16Group{}, &User{}, &v16Event{}, &v16EventAttendee{}, &v16Room{})
		},

		Rollback: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&v16GroupOld{}, &v16GroupMember{}, &v16GroupAdmin{})
		},
	}
}
