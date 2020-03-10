package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGormRepository_CreateUser(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupGormRepo(t, common)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		_, err := repo.CreateUser(mustNewUUIDV4(t), false)
		assert.NoError(t, err)
	})
}

func TestGormRepository_GetUser(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupGormRepo(t, common)
	user := mustMakeUser(t, repo, mustNewUUIDV4(t), false)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		_, err := repo.GetUser(user.ID)
		assert.NoError(t, err)
	})

}

func TestTraQRepository_GetAllUsers(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupTraQRepo(t)

	if users, err := repo.GetAllUsers(); assert.NoError(t, err) {
		assert.NotNil(t, users)
	}
}
