package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Migrate execute migrations
func Migrate(db *gorm.DB, tables []interface{}) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, Migrations())

	m.InitSchema(func(tx *gorm.DB) error {
		return tx.AutoMigrate(tables...)
	})
	return m.Migrate()
}
