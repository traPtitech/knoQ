package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type v12Event struct {
	Open bool `gorm:"default:false"`
}

type v12EventAttendee struct {
	UserID   uuid.UUID `gorm:"type:char(36); primaryKey"`
	EventID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	User     v12User   `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
	Schedule int
}

func (*v12EventAttendee) TableName() string {
	return "event_attendees"
}

type v12User struct {
	ID uuid.UUID `gorm:"type:char(36); primaryKey"`
}

func (*v12Event) TableName() string {
	return "events"
}

func (*v12User) TableName() string {
	return "users"
}

func v12() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "12",
		Migrate: func(db *gorm.DB) error {
			err := db.Migrator().AddColumn(&v12Event{}, "open")
			if err != nil {
				return err
			}
			return db.Migrator().CreateTable(&v12EventAttendee{})
		},
	}
}
