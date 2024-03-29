package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type v6currentRoom struct {
	Public bool
}

func (*v6currentRoom) TableName() string {
	return "rooms"
}

// v6 room.public -> room.verifed
func v6() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "6",
		Migrate: func(db *gorm.DB) error {
			return db.Migrator().RenameColumn(&v6currentRoom{}, "public", "verifed")
		},
	}
}
