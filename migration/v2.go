package migration

import (
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

type oldUser struct {
	ID    uuid.UUID `gorm:"type:char(36); primary_key"`
	Admin bool      `gorm:"not null"`
}

func (*oldUser) TableName() string {
	return "users"
}

type newUser struct {
	ID uuid.UUID `gorm:"type:char(36); primary_key"`
	// Admin アプリの管理者かどうか
	Admin      bool   `gorm:"not null"`
	IsTraq     bool   `gorm:"not null"`
	Token      string `gorm:"type:varbinary(64)"`
	IcalSecret string `gorm:"not null"`
}

func (*newUser) TableName() string {
	return "user_meta"
}

func v2() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "2",
		Migrate: func(db *gorm.DB) error {
			users := make([]*oldUser, 0)
			err := db.Find(&users).Error
			if err != nil {
				return err
			}
			err = db.CreateTable(&newUser{}).Error
			if err != nil {
				return err
			}
			for _, user := range users {
				err = db.Create(&newUser{
					ID:     user.ID,
					Admin:  user.Admin,
					IsTraq: true,
				}).Error
				if err != nil {
					return err
				}
			}
			return db.DropTableIfExists("users").Error
		},
	}
}
