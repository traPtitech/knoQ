package db

import (
	"fmt"

	"github.com/traPtitech/knoQ/migration"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// GormRepository implements domain
type GormRepository struct {
	db *gorm.DB
}

var tokenKey []byte

func (repo *GormRepository) Setup(host, user, password, database, key string) error {
	if host == "" {
		host = "mysql"
	}
	if user == "" {
		user = "root"
	}
	if password == "" {
		password = "password"
	}
	if database == "" {
		database = "knoQ"
	}
	if len(key) != 32 {
		panic("token key is not 32 words")
	}
	tokenKey = []byte(key)

	var err error
	repo.db, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=true&loc=Local", user, password, host, database),
		DefaultStringSize:         256,   // default size for string fields
		DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   false, // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
	}), &gorm.Config{})
	if err != nil {
		return err
	}

	return migration.Migrate(repo.db, tables)
}
