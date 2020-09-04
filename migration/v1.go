// Package migration migrate current struct
package migration

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

// v1 unique_index:idx_room_uniqueの削除
func v1() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "1",
		Migrate: func(db *gorm.DB) error {
			return db.
				Table("rooms").
				RemoveIndex("idx_room_unique").Error
		},
	}
}
