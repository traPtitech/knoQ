package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type v13User struct {
	ID          uuid.UUID `gorm:"type:char(36); primaryKey"`
	AccessToken string    `gorm:"type:varbinary(64)"`
}

func (*v13User) TableName() string {
	return "users"
}

type oldToken struct {
	UserID      uuid.UUID `gorm:"type:char(36); primaryKey"`
	AccessToken string    `gorm:"type:varbinary(64)"`
}

func (*oldToken) TableName() string {
	return "tokens"
}

func v13() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "13",
		Migrate: func(db *gorm.DB) error {
			// 1. ユーザーに Token カラムを追加
			if !db.Migrator().HasColumn(&v13User{}, "AccessToken") {
				if err := db.Migrator().AddColumn(&v13User{}, "AccessToken"); err != nil {
					return err
				}
			}

			// 2. 既存の AccessToken を User テーブルへ移行
			var tokens []oldToken
			if err := db.Find(&tokens).Error; err != nil {
				return err
			}
			for _, t := range tokens {
				if err := db.Model(&v13User{}).
					Where("id = ?", t.UserID).
					Update("access_token", t.AccessToken).Error; err != nil {
					return err
				}
			}

			// 3. Token テーブルを削除
			return db.Migrator().DropTable(&oldToken{})
		},
		Rollback: func(db *gorm.DB) error {
			type rollbackToken struct {
				UserID      uuid.UUID `gorm:"type:char(36); primaryKey"`
				AccessToken string    `gorm:"type:varbinary(64)"`
			}
			if err := db.Migrator().CreateTable(&rollbackToken{}); err != nil {
				return err
			}

			var users []v13User
			if err := db.Find(&users).Error; err != nil {
				return err
			}
			for _, u := range users {
				if u.AccessToken != "" {
					if err := db.Create(&rollbackToken{
						UserID:      u.ID,
						AccessToken: u.AccessToken,
					}).Error; err != nil {
						return err
					}
				}
			}

			if err := db.Migrator().DropColumn(&v13User{}, "AccessToken"); err != nil {
				return err
			}
			return nil
		},
	}
}
