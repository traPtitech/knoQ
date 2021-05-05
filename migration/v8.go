package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	newModel "github.com/traPtitech/knoQ/migration/v8"
	"gorm.io/gorm"
)

func v8() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "8",
		Migrate: func(db *gorm.DB) error {
			return db.AutoMigrate(newModel.Tables...)
		},
	}
}
