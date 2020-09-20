package repository

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	traQutils "github.com/traPtitech/traQ/utils"
)

func TestGormRepository_SaveUser(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupGormRepo(t, common)

	t.Run("success", func(t *testing.T) {
		userID := mustNewUUIDV4(t)
		_, err := repo.SaveUser(userID, false, true)
		assert.NoError(t, err)
	})

	t.Run("nil id", func(t *testing.T) {
		userID := uuid.Nil
		_, err := repo.SaveUser(userID, false, true)
		// mysql Error number 1364
		// Field 'id' doesn't have a default value
		assert.Error(t, err)
	})

	t.Run("already exists", func(t *testing.T) {
		userID := mustNewUUIDV4(t)
		_, err := repo.SaveUser(userID, false, true)
		assert.NoError(t, err)
		_, err = repo.SaveUser(userID, false, true)
		assert.EqualError(t, err, ErrAlreadyExists.Error())
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

func TestGormRepository_GetToken(t *testing.T) {
	t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)

	t.Run("Normal", func(t *testing.T) {
		tmp := traQutils.RandAlphabetAndNumberString(36)
		err := repo.ReplaceToken(user.ID, tmp)
		assert.NoError(t, err)
		token, err := repo.GetToken(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, tmp, token)
	})

}

func TestTraQRepository_GetAllUsers(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupTraQRepo(t, TraQv3)

	if users, err := repo.GetAllUsers(); assert.NoError(t, err) {
		assert.NotNil(t, users)
	}
}
