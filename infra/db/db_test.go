package db

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/traQ/utils/random"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	common = "common"
	ex     = "ex"

	dbUser = "root"
	dbPass = "password"
	dbHost = "localhost"
)

var (
	repositories = map[string]*GormRepository{}
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(err)
	}

	if err := pool.Client.Ping(); err != nil {
		panic(err)
	}

	resource, err := pool.Run("mariadb", "10.7", []string{fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", dbPass)})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := pool.Purge(resource); err != nil {
			panic(err)
		}
	}()

	var conn *sql.DB
	if err := pool.Retry(func() error {
		conn, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=true&loc=Local", dbUser, dbPass, dbHost, resource.GetPort("3306/tcp")))
		if err != nil {
			return err
		}
		return conn.Ping()
	}); err != nil {
		panic(err)
	}

	dbs := []string{common, ex}
	for _, v := range dbs {
		_, err = conn.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s%s`", "knoq-test-", v))
		if err != nil {
			panic(err)
		}
	}

	tokenKey = []byte(random.AlphaNumeric(32))

	for _, key := range dbs {
		db, err := gorm.Open(mysql.New(mysql.Config{
			DSN:                       fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", dbUser, dbPass, dbHost, resource.GetPort("3306/tcp"), "knoq-test-"+key),
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

func assertAndRequire(t *testing.T) (*assert.Assertions, *require.Assertions) {
	return assert.New(t), require.New(t)
}

func mustNewUUIDV4(t *testing.T) uuid.UUID {
	id, err := uuid.NewV4()
	require.NoError(t, err)
	return id
}

func setupRepo(t *testing.T, repo string) (*GormRepository, *assert.Assertions, *require.Assertions) {
	t.Helper()
	r, ok := repositories[repo]
	if !ok {
		t.FailNow()
	}
	assert, require := assertAndRequire(t)
	return r, assert, require
}

func setupRepoWithUser(t *testing.T, repo string) (*GormRepository, *assert.Assertions, *require.Assertions, *User) {
	t.Helper()
	r, assert, require := setupRepo(t, repo)
	user := mustMakeUser(t, r, false)
	return r, assert, require, user
}

func setupRepoWithUserGroup(t *testing.T, repo string) (*GormRepository, *assert.Assertions, *require.Assertions, *User, *Group) {
	t.Helper()
	r, assert, require := setupRepo(t, repo)
	group, user := mustMakeGroup(t, r, random.AlphaNumeric(10))
	return r, assert, require, user, group
}

func setupRepoWithUserRoom(t *testing.T, repo string) (*GormRepository, *assert.Assertions, *require.Assertions, *User, *Room) {
	t.Helper()
	r, assert, require := setupRepo(t, repo)
	room, user := mustMakeRoom(t, r, "here")
	return r, assert, require, user, room
}

func setupRepoWithUserGroupRoomEvent(t *testing.T, repo string) (*GormRepository, *assert.Assertions, *require.Assertions, *User, *Group, *Room, *Event) {
	t.Helper()
	r, assert, require := setupRepo(t, repo)

	event, group, room, user := mustMakeEvent(t, r, "event")
	return r, assert, require, user, group, room, event
}

func mustMakeUser(t *testing.T, repo *GormRepository, privilege bool) *User {
	t.Helper()
	user := User{
		Privilege: privilege,
	}
	err := repo.db.Create(&user).Error
	require.NoError(t, err)
	return &user
}

//func mustMakeUserBody(t *testing.T, repo *GormRepository, name, password string) *UserBody {
//t.Helper()
//user, err := saveUser(repo.db, userID uuid.UUID, privilege bool)
//require.NoError(t, err)
//return user
//}

// mustMakeGroup make group has no members
func mustMakeGroup(t *testing.T, repo *GormRepository, name string) (*Group, *User) {
	t.Helper()
	user := mustMakeUser(t, repo, false)
	params := WriteGroupParams{
		WriteGroupParams: domain.WriteGroupParams{
			Name:       name,
			Members:    nil,
			Admins:     []uuid.UUID{user.ID},
			JoinFreely: true,
		},
		CreatedBy: user.ID,
	}
	group, err := createGroup(repo.db, params)
	require.NoError(t, err)
	return group, user
}

//func mustAddGroupMember(t *testing.T, repo *GormRepository, groupID uuid.UUID, userID uuid.UUID) {
//t.Helper()
//err := repo.AddUserToGroup(groupID, userID)
//require.NoError(t, err)
//}

// mustMakeRoom make room. now -1h ~ now + 1h
func mustMakeRoom(t *testing.T, repo *GormRepository, place string) (*Room, *User) {
	t.Helper()

	user := mustMakeUser(t, repo, false)
	params := CreateRoomParams{
		WriteRoomParams: domain.WriteRoomParams{
			Place:     place,
			TimeStart: time.Now().Add(-1 * time.Hour),
			TimeEnd:   time.Now().Add(1 * time.Hour),
			Admins:    []uuid.UUID{user.ID},
		},
		CreatedBy: user.ID,
	}
	room, err := createRoom(repo.db, params)
	require.NoError(t, err)
	return room, user
}

func mustMakeTag(t *testing.T, repo *GormRepository, name string) *Tag {
	t.Helper()

	tag, err := createOrGetTag(repo.db, name)
	require.NoError(t, err)
	return tag
}

// mustMakeEvent make event. now ~ now + 1m
func mustMakeEvent(t *testing.T, repo *GormRepository, name string) (*Event, *Group, *Room, *User) {
	t.Helper()
	group, user := mustMakeGroup(t, repo, random.AlphaNumeric(10))
	room, _ := mustMakeRoom(t, repo, random.AlphaNumeric(10))

	params := WriteEventParams{
		WriteEventParams: domain.WriteEventParams{
			Name:          name,
			GroupID:       group.ID,
			RoomID:        room.ID,
			TimeStart:     time.Now(),
			TimeEnd:       time.Now().Add(1 * time.Minute),
			AllowTogether: true,
			Admins:        []uuid.UUID{user.ID},
			Tags: []domain.EventTagParams{
				{Name: "gin"},
			},
		},
		CreatedBy: user.ID,
	}

	event, err := createEvent(repo.db, params)
	require.NoError(t, err)
	return event, group, room, user
}
