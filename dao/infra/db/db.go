package db

import (
	"fmt"
	"os"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func (repo *GormRepository) Setup() error {
	host := os.Getenv("MARIADB_HOSTNAME")
	user := os.Getenv("MARIADB_USERNAME")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("MARIADB_PASSWORD")
	if password == "" {
		password = "password"
	}

	var err error
	repo.db, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=true&loc=Local", user, password, host, "knoQ"),
		DefaultStringSize:         256,   // default size for string fields
		DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
	}), &gorm.Config{})
	if err != nil {
		return err
	}

	err = repo.db.AutoMigrate(tables...)
	if err != nil {
		return err
	}
	return repo.db.SetupJoinTable(&Event{}, "Tags", &EventTag{})
}

func mustNewUUIDV4(t *testing.T) uuid.UUID {
	id, err := uuid.NewV4()
	require.NoError(t, err)
	return id
}
