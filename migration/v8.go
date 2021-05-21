package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	newModel "github.com/traPtitech/knoQ/migration/v8"
	"gorm.io/gorm"
)

// v8 created_by -> created_by_refer
func v8() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "8",
		Migrate: func(db *gorm.DB) error {
			err := db.Migrator().RenameColumn(&newModel.Event{}, "created_by", "created_by_refer")
			if err != nil {
				return err
			}
			return db.Migrator().RenameColumn(&newModel.Group{}, "created_by", "created_by_refer")

		},
	}
}
