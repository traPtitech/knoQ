package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGormRepository_SaveUser(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupGormRepo(t, common)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		_, err := repo.SaveUser(false)
		assert.NoError(t, err)
	})
}

func TestGormRepository_GetUser(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupGormRepo(t, common)
	user := mustMakeUserMeta(t, repo, false)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		_, err := repo.GetUser(user.ID)
		assert.NoError(t, err)
	})

}

func TestGormRepository_ReplaceToken(t *testing.T) {
	t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)

	t.Run("Normal", func(t *testing.T) {
		err := repo.ReplaceToken(user.ID, "0123456789abcdefghijklmn")
		assert.NoError(t, err)
		token, err := repo.GetToken(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, "0123456789abcdefghijklmn", token)
	})

}

func TestTraQRepository_GetAllUsers(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupTraQRepo(t, TraQv3)

	if users, err := repo.GetAllUsers(); assert.NoError(t, err) {
		assert.NotNil(t, users)
	}
}
