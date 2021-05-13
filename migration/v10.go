package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// delete sessions
func v10() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "10",
		Migrate: func(db *gorm.DB) error {
			return db.Migrator().DropTable("sessions")
		},
	}
}
