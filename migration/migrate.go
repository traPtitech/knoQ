package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Migrate execute migrations
func Migrate(db *gorm.DB, tables []interface{}) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, Migrations())

	m.InitSchema(func(tx *gorm.DB) error {
		mv1 := gormigrate.New(tx, gormigrate.DefaultOptions, []*gormigrate.Migration{
			{
				ID:      "assume existing DB",
				Migrate: v1().Migrate,
			},
		})
		err := mv1.Migrate()
		if err != nil {
			return err
		}
		err = tx.AutoMigrate(tables...)
		if err != nil {
			return err
		}

		return nil
	})
	return m.Migrate()
}
