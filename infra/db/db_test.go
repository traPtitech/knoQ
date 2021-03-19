package db

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/traQ/migration"
	"github.com/traPtitech/traQ/utils/random"
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

func mustMakeUser(t *testing.T, repo *GormRepository, privilege bool) *User {
	t.Helper()
	userID := mustNewUUIDV4(t)
	user, err := saveUser(repo.db, &User{
		ID:        userID,
		Privilege: privilege,
	})
	require.NoError(t, err)
	return user
}

//func mustMakeUserBody(t *testing.T, repo *GormRepository, name, password string) *UserBody {
//t.Helper()
//user, err := saveUser(repo.db, userID uuid.UUID, privilege bool)
//require.NoError(t, err)
//return user
//}

// mustMakeGroup make group has no members
func mustMakeGroup(t *testing.T, repo *GormRepository, name string, createdBy uuid.UUID) *Group {
	t.Helper()
	params := writeGroupParams{
		WriteGroupParams: domain.WriteGroupParams{
			Name:       name,
			Members:    nil,
			JoinFreely: true,
		},
		CreatedBy: createdBy,
	}
	group, err := createGroup(repo.db, params)
	require.NoError(t, err)
	return group
}

//func mustAddGroupMember(t *testing.T, repo *GormRepository, groupID uuid.UUID, userID uuid.UUID) {
//t.Helper()
//err := repo.AddUserToGroup(groupID, userID)
//require.NoError(t, err)
//}

// mustMakeRoom make room. now -1h ~ now + 1h
func mustMakeRoom(t *testing.T, repo *GormRepository, place string) *Room {
	t.Helper()
	params := writeRoomParams{
		WriteRoomParams: domain.WriteRoomParams{
			Place:     place,
			TimeStart: time.Now().Add(-1 * time.Hour),
			TimeEnd:   time.Now().Add(1 * time.Hour),
		},
	}
	room, err := createRoom(repo.db, params)
	require.NoError(t, err)
	return room
}

func mustMakeTag(t *testing.T, repo *GormRepository, name string) *Tag {
	t.Helper()

	tag, err := createTag(repo.db, name)
	require.NoError(t, err)
	return tag
}

// mustMakeEvent make event. now ~ now + 1m
func mustMakeEvent(t *testing.T, repo *GormRepository, name string, userID uuid.UUID) (*Event, *Group, *Room) {
	t.Helper()
	group := mustMakeGroup(t, repo, random.AlphaNumeric(10), userID)
	room := mustMakeRoom(t, repo, random.AlphaNumeric(10))

	params := writeEventParams{
		WriteEventParams: domain.WriteEventParams{
			Name:      name,
			GroupID:   group.ID,
			RoomID:    room.ID,
			TimeStart: time.Now(),
			TimeEnd:   time.Now().Add(1 * time.Minute),
		},
		CreatedBy: userID,
	}

	event, err := createEvent(repo.db, params)
	require.NoError(t, err)
	return event, group, room
}
