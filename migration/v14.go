package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type v14User struct {
	ID           uuid.UUID `gorm:"type:char(36); primaryKey"`
	ProviderName string    `gorm:"not null"`
}

func (*v14User) TableName() string {
	return "users"
}

type oldProvider struct {
	UserID uuid.UUID `gorm:"type:char(36); primaryKey"`
	Issuer string    `gorm:"not null"`
}

func (*oldProvider) TableName() string {
	return "providers"
}

func v14() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "14",
		Migrate: func(db *gorm.DB) error {
			// 1. カラム追加（存在チェック付き）
			if !db.Migrator().HasColumn(&v14User{}, "ProviderName") {
				if err := db.Migrator().AddColumn(&v14User{}, "ProviderName"); err != nil {
					return err
				}
			}

			// 2. providers から issuer を users にコピー
			var providers []oldProvider
			if err := db.Find(&providers).Error; err != nil {
				return err
			}
			for _, p := range providers {
				if err := db.Model(&v14User{}).
					Where("id = ?", p.UserID).
					Update("provider_name", p.Issuer).Error; err != nil {
					return err
				}
			}

			// 3. providers テーブル削除
			return db.Migrator().DropTable(&oldProvider{})
		},
		Rollback: func(db *gorm.DB) error {
			// 任意の rollback 実装（必要に応じて）
			type rollbackProvider struct {
				UserID uuid.UUID `gorm:"type:char(36); primaryKey"`
				Issuer string    `gorm:"not null"`
			}
			if err := db.Migrator().CreateTable(&rollbackProvider{}); err != nil {
				return err
			}
			var users []v14User
			if err := db.Find(&users).Error; err != nil {
				return err
			}
			for _, u := range users {
				if u.ProviderName != "" {
					if err := db.Create(&rollbackProvider{
						UserID: u.ID,
						Issuer: u.ProviderName,
					}).Error; err != nil {
						return err
					}
				}
			}
			return db.Migrator().DropColumn(&v14User{}, "ProviderName")
		},
	}
}
