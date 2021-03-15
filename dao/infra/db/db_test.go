package db

import (
	"fmt"
	"os"
	"testing"

	"github.com/traPtitech/traQ/migration"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	common = "common"
	ex     = "ex"
)

var (
	repositories = map[string]*GormRepository{}
)

func TestMain(m *testing.M) {
	const (
		user     = "root"
		password = "password"
		host     = "localhost"
	)

	dbs := []string{
		common,
		ex,
	}
	if err := migration.CreateDatabasesIfNotExists("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/?charset=utf8mb4&parseTime=true&loc=Local", user, password, host), "knoq-test-", dbs...); err != nil {
		panic(err)
	}

	for _, key := range dbs {
		db, err := gorm.Open(mysql.New(mysql.Config{
			DSN:                       fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=true&loc=Local", user, password, host, "knoq-test-"+key),
			DefaultStringSize:         256,   // default size for string fields
			DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
			DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
			DontSupportRenameColumn:   true,  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
			SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
		}), &gorm.Config{})
		if err != nil {
			panic(err)
		}

		sqlDB, _ := db.DB()
		sqlDB.SetMaxOpenConns(20)

		if err := db.Migrator().DropTable(tables...); err != nil {
			panic(err)
		}
		if err := db.Migrator().AutoMigrate(tables...); err != nil {
			panic(err)
		}
		repo := GormRepository{
			db: db,
		}
		repositories[key] = &repo
	}

	code := m.Run()
	os.Exit(code)
}
