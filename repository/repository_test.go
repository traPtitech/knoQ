// Package repository is
package repository

import (
	"fmt"
	"os"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/traQ/migration"
)

const (
	common = "common"
	ex     = "ex"
)

var (
	repositories = map[string]*GormRepository{}
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
	if err := migration.CreateDatabasesIfNotExists("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/?charset=utf8mb4&parseTime=true", user, password, host), "room-test-", dbs...); err != nil {
		panic(err)
	}

	for _, key := range dbs {
		db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=true", user, password, host, "room-test-"+key))
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

		code := m.Run()
		for _, v := range repositories {
			v.DB.Close()
		}
		os.Exit(code)
	}
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
	user := mustMakeUser(t, r, mustNewUUIDV4(t), false)
	return r, assert, require, user
}

func mustMakeUser(t *testing.T, repo UserRepository, userID uuid.UUID, admin bool) *User {
	t.Helper()
	user, err := repo.CreateUser(userID, admin)
	require.NoError(t, err)
	return user
}

// mustMakeGroup make group has no members
func mustMakeGroup(t *testing.T, repo GroupRepository, name string, createdBy uuid.UUID) *Group {
	t.Helper()
	params := WriteGroupParams{
		Name:      name,
		Members:   nil,
		CreatedBy: createdBy,
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

func setupTraQRepo(t *testing.T) (*TraQRepository, *assert.Assertions, *require.Assertions) {
	t.Helper()
	repo := &TraQRepository{
		APIRepository: APIRepository{
			BaseURL: "https://q.trap.jp/api/v3",
		},
		Token: os.Getenv("TRAQ_AUTH"),
	}
	assert, require := assertAndRequire(t)
	return repo, assert, require

}
