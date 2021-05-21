package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	newModel "github.com/traPtitech/knoQ/migration/v8"
)

// v9 足りないカラムの追加、キーの作成
func v9() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "9",
		Migrate: func(db *gorm.DB) error {
			return db.AutoMigrate(newModel.Tables...)
		},
	}
}
