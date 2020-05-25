// Package repository is
package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/traQ/migration"
	traQutils "github.com/traPtitech/traQ/utils"
)

const (
	common = "common"
	ex     = "ex"
)

var (
	repositories = map[string]*GormRepository{}
	authCookie   *http.Cookie
)

func TestMain(m *testing.M) {
	host := os.Getenv("MARIADB_HOSTNAME")
	user := os.Getenv("MARIADB_USERNAME")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("MARIADB_PASSWORD")
	if password == "" {
		password = "password"
	}

	dbs := []string{
		common,
		ex,
	}
	if err := migration.CreateDatabasesIfNotExists("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/?charset=utf8mb4&parseTime=true&loc=Local", user, password, host), "room-test-", dbs...); err != nil {
		panic(err)
	}

	for _, key := range dbs {
		db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=true&loc=Local", user, password, host, "room-test-"+key))
		if err != nil {
			panic(err)
		}
		db.DB().SetMaxOpenConns(20)
		if err := db.DropTableIfExists(tables...).Error; err != nil {
			panic(err)
		}
		if err := initDB(db); err != nil {
			panic(err)
		}
		repo := GormRepository{
			DB: db,
		}
		repositories[key] = &repo
	}

	// traQ
	// 1. /login
	authentication := struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}{
		Name:     "traq",
		Password: "traq",
	}
	body, err := json.Marshal(authentication)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, "http://localhost:3000/api/v3/login", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	if res.StatusCode >= 300 {
		panic("unexpected status code")
	}
	// 2. get Set-Cookie
	cookies := res.Cookies()
	for _, v := range cookies {
		if v.Name == "r_session" {
			authCookie = v
			break
		}
	}

	code := m.Run()
	for _, v := range repositories {
		v.DB.Close()
	}
	os.Exit(code)
}

func assertAndRequire(t *testing.T) (*assert.Assertions, *require.Assertions) {
	return assert.New(t), require.New(t)
}

func mustNewUUIDV4(t *testing.T) uuid.UUID {
	id, err := uuid.NewV4()
	require.NoError(t, err)
	return id
}

func setupGormRepo(t *testing.T, repo string) (*GormRepository, *assert.Assertions, *require.Assertions) {
	t.Helper()
	r, ok := repositories[repo]
	if !ok {
		t.FailNow()
	}
	assert, require := assertAndRequire(t)
	return r, assert, require
}

func setupGormRepoWithUser(t *testing.T, repo string) (*GormRepository, *assert.Assertions, *require.Assertions, *User) {
	t.Helper()
	r, assert, require := setupGormRepo(t, repo)
	user := mustMakeUser(t, r, false)
	return r, assert, require, user
}

func mustMakeUser(t *testing.T, repo UserRepository, admin bool) *User {
	t.Helper()
	user, err := repo.CreateUser(admin)
	require.NoError(t, err)
	return user
}

// mustMakeGroup make group has no members
func mustMakeGroup(t *testing.T, repo GroupRepository, name string, createdBy uuid.UUID) *Group {
	t.Helper()
	params := WriteGroupParams{
		Name:       name,
		Members:    nil,
		JoinFreely: true,
		CreatedBy:  createdBy,
	}
	group, err := repo.CreateGroup(params)
	require.NoError(t, err)
	return group
}

func mustAddGroupMember(t *testing.T, repo GroupRepository, groupID uuid.UUID, userID uuid.UUID) {
	t.Helper()
	err := repo.AddUserToGroup(groupID, userID)
	require.NoError(t, err)
}

// mustMakeRoom make room. now -1h ~ now + 1h
func mustMakeRoom(t *testing.T, repo RoomRepository, place string) *Room {
	t.Helper()
	params := WriteRoomParams{
		Place:     place,
		TimeStart: time.Now().Add(-1 * time.Hour),
		TimeEnd:   time.Now().Add(1 * time.Hour),
	}
	room, err := repo.CreateRoom(params)
	require.NoError(t, err)
	return room
}

func mustMakeTag(t *testing.T, repo TagRepository, name string) *Tag {
	t.Helper()

	tag, err := repo.CreateOrGetTag(name)
	require.NoError(t, err)
	return tag
}

// mustMakeEvent make event. now ~ now + 1m
func mustMakeEvent(t *testing.T, repo Repository, name string, userID uuid.UUID) (*Event, *Group, *Room) {
	t.Helper()
	group := mustMakeGroup(t, repo, traQutils.RandAlphabetAndNumberString(10), userID)
	room := mustMakeRoom(t, repo, traQutils.RandAlphabetAndNumberString(10))

	params := WriteEventParams{
		Name:      name,
		GroupID:   group.ID,
		RoomID:    room.ID,
		TimeStart: time.Now(),
		TimeEnd:   time.Now().Add(1 * time.Minute),
		CreatedBy: userID,
	}

	event, err := repo.CreateEvent(params)
	require.NoError(t, err)
	return event, group, room
}

func setupTraQRepo(t *testing.T, version TraQVersion) (*TraQRepository, *assert.Assertions, *require.Assertions) {
	t.Helper()
	repo := &TraQRepository{
		Version: version,
		Host:    "http://localhost:3000/api",
	}
	repo.NewRequest = func(method, url string, body io.Reader) (*http.Request, error) {
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return nil, err
		}
		req.AddCookie(authCookie)
		return req, nil
	}
	assert, require := assertAndRequire(t)
	return repo, assert, require
}

func setupTraPGroupRepo(t *testing.T, version TraQVersion) (*TraPGroupRepository, *assert.Assertions, *require.Assertions) {
	repo := new(TraPGroupRepository)
	repoTraQ, assert, require := setupTraQRepo(t, version)
	repo.TraQRepository = *repoTraQ
	return repo, assert, require
}
