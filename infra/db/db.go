package db

import (
	"time"

	gomysql "github.com/go-sql-driver/mysql"
	"github.com/traPtitech/knoQ/infra"
	"github.com/traPtitech/knoQ/migration"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormRepository implements domain
type GormRepository struct {
	db       *gorm.DB
	traqRepo infra.TraqRepository
}

func (repo *GormRepository) GetTraqRepository() infra.TraqRepository {
	return repo.traqRepo
}

var tokenKey []byte

func (repo *GormRepository) Setup(host, user, password, database, port, key, logLevel string, loc *time.Location, traqRepo infra.TraqRepository) error {
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

	var err error
	repo.db, err = gorm.Open(
		mysql.New(mysql.Config{
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
		&gorm.Config{
			Logger: logger.Default.LogMode(loglevel),
		})
	if err != nil {
		return err
	}

	repo.traqRepo = traqRepo
	return migration.Migrate(repo.db, tables)
}
