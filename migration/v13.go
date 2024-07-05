package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type v13Post struct {
	MessageID uuid.UUID `gorm:"type:char(36); primaryKey"`
	EventID   uuid.UUID `gorm:"type:char(36); not null"`
}

func (*v13Post) TableName() string {
	return "posts"
}

func v13() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "13",
		Migrate: func(db *gorm.DB) error {
			return db.Migrator().CreateTable(&v13Post{})
		},
	}
}
