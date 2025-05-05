package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type v17User struct {
	Privilege bool
}

func (*v17User) TableName() string {
	return "users"
}

func v17() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "17",
		Migrate: func(db *gorm.DB) error {
			if db.Migrator().HasColumn(&v17User{}, "privilege") {
				return db.Migrator().RenameColumn(&v17User{}, "privilege", "privileged")
			}
			return nil
		},
	}
}
