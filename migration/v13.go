package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID         uuid.UUID `gorm:"type:char(36); primaryKey"`
	Privilege  bool      `gorm:"not null"`
	State      int
	IcalSecret string `gorm:"not null"`
	Issuer     string `gorm:"not null"`
	Subject    string
}

type Provider struct {
	UserID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	Issuer  string    `gorm:"not null"`
	Subject string
}

func v13() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "13",
		Migrate: func(db *gorm.DB) error {
			// Step 1: Add Issuer and Subject columns to the User table
			if err := db.Migrator().AddColumn(&User{}, "Issuer"); err != nil {
				return err
			}
			if err := db.Migrator().AddColumn(&User{}, "Subject"); err != nil {
				return err
			}

			// Step 2: Migrate data from Provider to User
			providers := make([]*Provider, 0)
			if err := db.Find(&providers).Error; err != nil {
				return err
			}

			for _, provider := range providers {
				if err := db.Model(&User{}).Where("id = ?", provider.UserID).Updates(map[string]interface{}{
					"Issuer":  provider.Issuer,
					"Subject": provider.Subject,
				}).Error; err != nil {
					return err
				}
			}

			// Step 3: Drop the Provider table
			if err := db.Migrator().DropTable(&Provider{}); err != nil {
				return err
			}

			return nil
		},
	}
}
