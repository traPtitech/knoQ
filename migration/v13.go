package migration

import (
	"time"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type embeddedToken struct {
	AccessToken  string `gorm:"type:varbinary(64)"`
	TokenType    string
	RefreshToken string
	Expiry       time.Time
}

type v13newUser struct {
	ID         uuid.UUID `gorm:"type:char(36); primaryKey"`
	Privilege  bool      `gorm:"not null"`
	State      int
	IcalSecret string `gorm:"not null"`
	Issuer     string `gorm:"not null"`
	Subject    string
	*embeddedToken
}

func (*v13newUser) TableName() string {
	return "users"
}

type v13currentProvider struct {
	UserID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	Issuer  string    `gorm:"not null"`
	Subject string
}

func (*v13currentProvider) TableName() string {
	return "providers"
}

type v13currentToken struct {
	UserID uuid.UUID `gorm:"type:char(36); primaryKey"`
	*embeddedToken
}

func (*v13currentToken) TableName() string {
	return "tokens"
}

func v13() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "13",
		Migrate: func(db *gorm.DB) error {
			// Step 1: Add Issuer and Subject columns to the User table
			if err := db.Migrator().AddColumn(&v13newUser{}, "Issuer"); err != nil {
				return err
			}
			if err := db.Migrator().AddColumn(&v13newUser{}, "Subject"); err != nil {
				return err
			}

			// Step 2: Migrate data from Provider to User
			providers := make([]*v13currentProvider, 0)
			if err := db.Find(&providers).Error; err != nil {
				return err
			}

			for _, provider := range providers {
				if err := db.Model(&v13newUser{}).Where("id = ?", provider.UserID).Updates(map[string]interface{}{
					"Issuer":  provider.Issuer,
					"Subject": provider.Subject,
				}).Error; err != nil {
					return err
				}
			}

			// Step 3: Drop the Provider table
			if err := db.Migrator().DropTable(&v13currentProvider{}); err != nil {
				return err
			}

			// Step 1: Add Token fields to the User table
			if err := db.Migrator().AddColumn(&v13newUser{}, "AccessToken"); err != nil {
				return err
			}
			if err := db.Migrator().AddColumn(&v13newUser{}, "TokenType"); err != nil {
				return err
			}
			if err := db.Migrator().AddColumn(&v13newUser{}, "RefreshToken"); err != nil {
				return err
			}
			if err := db.Migrator().AddColumn(&v13newUser{}, "Expiry"); err != nil {
				return err
			}

			// Step 2: Migrate data from Token to User
			tokens := make([]*v13currentToken, 0)
			if err := db.Find(&tokens).Error; err != nil {
				return err
			}

			for _, token := range tokens {
				if err := db.Model(&v13newUser{}).Where("id = ?", token.UserID).Updates(map[string]interface{}{
					"AccessToken":  token.AccessToken,
					"TokenType":    token.TokenType,
					"RefreshToken": token.RefreshToken,
					"Expiry":       token.Expiry,
				}).Error; err != nil {
					return err
				}
			}

			// Step 3: Drop the Token table
			if err := db.Migrator().DropTable(&v13currentToken{}); err != nil {
				return err
			}

			return nil
		},
	}
}
