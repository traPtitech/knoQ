package db

import (
	"time"

	gomysql "github.com/go-sql-driver/mysql"
	"github.com/traPtitech/knoQ/migration"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormRepository implements domain
type GormRepository struct {
	db *gorm.DB
}

var tokenKey []byte

func (repo *GormRepository) Setup(host, user, password, database, port, key, logLevel string, loc *time.Location) error {
	loglevel := func() logger.LogLevel {
		switch logLevel {
		case "slient":
			return logger.Silent
		case "error":
			return logger.Error
		case "info":
			return logger.Info
		case "warn":
			return logger.Warn
		default:
			return logger.Silent
		}
	}()
	if len(key) != 32 {
		panic("token key is not 32 words")
	}
	tokenKey = []byte(key)

	d := dialector{
		Dialector: mysql.New(mysql.Config{
			DSNConfig: &gomysql.Config{
				User:                 user,
				Passwd:               password,
				Net:                  "tcp",
				Addr:                 host + ":" + port,
				DBName:               database,
				Loc:                  loc,
				AllowNativePasswords: true,
				ParseTime:            true,
			},
			DefaultStringSize:         256,   // default size for string fields
			DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
			DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
			DontSupportRenameColumn:   false, // `change` when rename column, rename column not supported before MySQL 8, MariaDB
			SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
		}),
	}

	var err error
	repo.db, err = gorm.Open(d, &gorm.Config{
		Logger: logger.Default.LogMode(loglevel),
	})
	if err != nil {
		return err
	}

	return migration.Migrate(repo.db, tables)
}

// dialector with custom error handling
type dialector struct {
	gorm.Dialector
	gorm.SavePointerDialectorInterface
}

var (
	_ gorm.Dialector       = (*dialector)(nil)
	_ gorm.ErrorTranslator = (*dialector)(nil)
)

// override Translate(err error) error
func (d dialector) Translate(err error) error {
	if translater, ok := d.Dialector.(gorm.ErrorTranslator); ok {
		err = translater.Translate(err)
	}

	return defaultErrorHandling(err)
}
