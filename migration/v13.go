package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func v13() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "13",
		Migrate: func(db *gorm.DB) error {
			return db.Exec("ALTER TABLE events DROP CONSTRAINT UNIQUE(name, time_start, time_end)").Error
		},
	}
}
