package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func v5() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "5",
		Migrate: func(db *gorm.DB) error {
			return db.Migrator().RenameTable("group_users", "group_members")
		},
	}
}
