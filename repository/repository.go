// Package repository is
package repository

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"room/migration"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jwt"

	// mysql
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// 新たにモデル(テーブル)を定義した場合はここに追加する事
var tables = []interface{}{
	User{},
	Room{},
	Group{},
	Event{},
	Tag{},
	EventTag{},
	GroupUsers{},
}

type Repository interface {
	UserRepository
	GroupRepository
	RoomRepository
	EventRepository
	TagRepository
}

// GormRepository implements Repository interface
type GormRepository struct {
	DB *gorm.DB
}

type TraQVersion int64

const (
	TraQv1 TraQVersion = iota
	TraQv3
)

type TraQRepository struct {
	Version TraQVersion
	Host    string
	Token   string
	// required
	NewRequest func(method string, url string, body io.Reader) (*http.Request, error)
}

// TraPGroupRepository traPメンバー全体をグループとして扱う
type TraPGroupRepository struct {
	TraQRepository
}
type GoogleAPIRepository struct {
	Config     *jwt.Config
	Client     *http.Client
	CalendarID string
}

var (
	MARIADB_HOSTNAME = os.Getenv("MARIADB_HOSTNAME")
	MARIADB_DATABASE = os.Getenv("MARIADB_DATABASE")
	MARIADB_USERNAME = os.Getenv("MARIADB_USERNAME")
	MARIADB_PASSWORD = os.Getenv("MARIADB_PASSWORD")

	DB        *gorm.DB
	logger, _ = zap.NewDevelopment()
)

func (repo *GoogleAPIRepository) Setup() {
	repo.Client = repo.Config.Client(oauth2.NoContext)
}

func (repo *TraQRepository) getBaseURL() string {
	var traQEndPointVersion = [2]string{
		"/1.0",
		"/v3",
	}

	return repo.Host + traQEndPointVersion[repo.Version]
}

// DefaultNewRequest set Authorization Header
func (repo *TraQRepository) DefaultNewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	if repo.Token == "" {
		return nil, ErrForbidden
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+repo.Token)

	return req, nil
}

func (repo *TraQRepository) getRequest(path string) ([]byte, error) {
	req, err := repo.NewRequest(http.MethodGet, repo.getBaseURL()+path, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if err := judgeStatusCode(res.StatusCode); err != nil {
		return nil, err
	}
	return ioutil.ReadAll(res.Body)
}

func (repo *TraQRepository) postRequest(path string, body []byte) ([]byte, error) {
	req, err := repo.NewRequest(http.MethodPost, repo.getBaseURL()+path, bytes.NewBuffer(body))

	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if err := judgeStatusCode(res.StatusCode); err != nil {
		return nil, err
	}
	return ioutil.ReadAll(res.Body)
}

func judgeStatusCode(code int) error {
	if code >= 300 {
		// TODO consider 300
		switch code {
		case 401:
			return ErrForbidden
		case 403:
			return ErrForbidden
		case 404:
			return ErrNotFound
		default:
			return errors.New(http.StatusText(code))
		}
	}
	return nil
}

// SetupDatabase set up DB and crate tables
func SetupDatabase() (*gorm.DB, error) {
	var err error
	//tmp
	if MARIADB_HOSTNAME == "" {
		MARIADB_HOSTNAME = ""
	}
	if MARIADB_DATABASE == "" {
		MARIADB_DATABASE = "room"
	}
	if MARIADB_USERNAME == "" {
		MARIADB_USERNAME = "root"
	}

	if MARIADB_PASSWORD == "" {
		MARIADB_PASSWORD = "password"
	}

	// データベース接続
	DB, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", MARIADB_USERNAME, MARIADB_PASSWORD, MARIADB_HOSTNAME, MARIADB_DATABASE))
	if err != nil {
		return DB, err
	}
	if err := initDB(DB); err != nil {
		return DB, err
	}
	return DB, nil
}

// initDB データベースのスキーマを更新
func initDB(db *gorm.DB) error {
	// gormのエラーの上書き
	gorm.ErrRecordNotFound = ErrNotFound
	// db.LogMode(true)
	if err := migration.Migrate(db, tables); err != nil {
		return err
	}
	return nil
}
