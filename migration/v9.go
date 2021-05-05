package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	newModel "github.com/traPtitech/knoQ/migration/v8"
)

// rename table group_users -> group_members
func v9() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "9",
		Migrate: func(db *gorm.DB) error {
			return db.AutoMigrate(newModel.Tables...)
		},
	}
}
