package migration

import (
	"time"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// v7newOauth2Token is oauth2.Token
type v7Oauth2Token struct {
	// AccessToken is the token that authorizes and authenticates
	// the requests.
	AccessToken string `gorm:"type:varbinary(64)"`

	// TokenType is the type of token.
	// The Type method returns either this or "Bearer", the default.
	TokenType string

	// RefreshToken is a token that's used by the application
	// (as opposed to the user) to refresh the access token
	// if it expires.
	RefreshToken string

	// Expiry is the optional expiration time of the access token.
	//
	// If zero, TokenSource implementations will reuse the same
	// token forever and RefreshToken or equivalent
	// mechanisms for that TokenSource will not be used.
	Expiry time.Time
}

type v7newToken struct {
	UserID uuid.UUID `gorm:"type:char(36); primaryKey"`

	*v7Oauth2Token
}

func (*v7newToken) TableName() string {
	return "tokens"
}

type v7newProvider struct {
	UserID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	Issuer  string    `gorm:"not null"`
	Subject string
}

func (*v7newProvider) TableName() string {
	return "providers"
}

type v7newUser struct {
	ID uuid.UUID `gorm:"type:char(36); primaryKey"`
	// アプリの管理者かどうか
	Privilege  bool `gorm:"<-:create; not null"` // Do not update
	State      int
	IcalSecret string        `gorm:"not null"`
	Provider   v7newProvider `gorm:"foreignKey:UserID; constraint:OnDelete:CASCADE;"`
	Token      v7newToken    `gorm:"foreignKey:UserID; constraint:OnDelete:CASCADE;"`
}

func (*v7newUser) TableName() string {
	return "user"
}

type v7newUserBody struct {
	ID          uuid.UUID `gorm:"type:char(36); primaryKey;"`
	Name        string    `gorm:"type:varchar(32);"`
	DisplayName string    `gorm:"type:varchar(32);"`
	Icon        string
	User        v7newUser `gorm:"->; foreignKey:ID; constraint:OnDelete:CASCADE;" cvt:"->"`
}

func (*v7newUserBody) TableName() string {
	return "user_body"
}

type v7currentUserMeta struct {
	ID uuid.UUID `gorm:"type:char(36); primary_key"`
	// Admin アプリの管理者かどうか
	Admin      bool   `gorm:"not null"`
	IsTraq     bool   `gorm:"not null"`
	Token      string `gorm:"type:varbinary(64)"`
	IcalSecret string `gorm:"not null"`
}

func (*v7currentUserMeta) TableName() string {
	return "user_meta"
}

func Convv7currentUserMetaTov7newUser(src v7currentUserMeta) (dst v7newUser) {
	dst.ID = src.ID
	dst.IcalSecret = src.IcalSecret
	dst.Privilege = src.Admin
	dst.State = 1
	return
}

func ConvSPv7currentUserMetaToSPv7newUser(src []*v7currentUserMeta) (dst []*v7newUser) {
	dst = make([]*v7newUser, len(src))
	for i := range src {
		dst[i] = new(v7newUser)
		(*dst[i]) = Convv7currentUserMetaTov7newUser((*src[i]))
	}
	return
}

func v7() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "7",
		Migrate: func(db *gorm.DB) error {
			cums := make([]*v7currentUserMeta, 0)
			err := db.Find(&cums).Error
			if err != nil {
				return err
			}

			db.Migrator().CreateTable(&v7newUser{})
			db.Migrator().CreateTable(&v7newUserBody{})
			db.Migrator().CreateTable(&v7newToken{})
			db.Migrator().CreateTable(&v7newProvider{})

			nus := ConvSPv7currentUserMetaToSPv7newUser(cums)
			err = db.Create(&nus).Error
			if err != nil {
				return err
			}
			return db.Migrator().DropTable(&v7currentUserMeta{})
		},
	}
}
