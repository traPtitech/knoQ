package repository

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	traQutils "github.com/traPtitech/traQ/utils"
)

func TestGormRepository_CreateGroup(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupGormRepo(t, common)
	user := mustMakeUser(t, repo, mustNewUUIDV4(t), false)

	params := WriteGroupParams{
		Name:      traQutils.RandAlphabetAndNumberString(20),
		Members:   []uuid.UUID{user.ID, mustNewUUIDV4(t)},
		CreatedBy: user.ID,
	}

	if group, err := repo.CreateGroup(params); assert.NoError(t, err) {
		assert.NotNil(t, group)
		assert.Equal(t, params.Members[0], group.Members[0].ID)
		assert.Equal(t, 1, len(group.Members))
	}
}

func TestGormRepository_UpdateGroup(t *testing.T) {
	t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)
	group := mustMakeGroup(t, repo, traQutils.RandAlphabetAndNumberString(20), user.ID)

	params := WriteGroupParams{
		Name:      group.Name,
		Members:   []uuid.UUID{user.ID, mustNewUUIDV4(t)},
		CreatedBy: user.ID,
	}

	if group, err := repo.UpdateGroup(group.ID, params); assert.NoError(t, err) {
		assert.NotNil(t, group)
		assert.Equal(t, params.Members[0], group.Members[0].ID)
		assert.Equal(t, 1, len(group.Members))
	}
}

func TestGormRepository_AddUserToGroup(t *testing.T) {
	t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)
	group := mustMakeGroup(t, repo, traQutils.RandAlphabetAndNumberString(20), user.ID)

	t.Run("Add existing user", func(t *testing.T) {
		t.Parallel()
		err := repo.AddUserToGroup(group.ID, user.ID)
		assert.NoError(t, err)
	})

	t.Run("Add not existing user", func(t *testing.T) {
		t.Parallel()
		err := repo.AddUserToGroup(group.ID, mustNewUUIDV4(t))
		assert.EqualError(t, err, ErrNotFound.Error())
	})

	t.Run("Add already exists user", func(t *testing.T) {
		t.Parallel()
		user := mustMakeUser(t, repo, mustNewUUIDV4(t), false)
		err := repo.AddUserToGroup(group.ID, user.ID)
		assert.NoError(t, err)
		err = repo.AddUserToGroup(group.ID, user.ID)
		// association_autoupdate:false;association_autocreate:false" の影響?
		assert.NoError(t, err)
	})
}

func TestGormRepository_DeleteGroup(t *testing.T) {
	t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)
	group := mustMakeGroup(t, repo, traQutils.RandAlphabetAndNumberString(20), user.ID)

	t.Run("Delete existing group", func(t *testing.T) {
		t.Parallel()
		err := repo.DeleteGroup(group.ID)
		assert.NoError(t, err)
	})

	t.Run("Delete not existing group", func(t *testing.T) {
		t.Parallel()
		err := repo.DeleteGroup(mustNewUUIDV4(t))
		assert.EqualError(t, err, ErrNotFound.Error())
	})
}

func TestGormRepository_DeleteUserInGroup(t *testing.T) {
	t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)
	group := mustMakeGroup(t, repo, traQutils.RandAlphabetAndNumberString(20), user.ID)
	mustAddGroupMember(t, repo, group.ID, user.ID)

	t.Run("Delete existing member in group", func(t *testing.T) {
		t.Parallel()
		err := repo.DeleteUserInGroup(group.ID, user.ID)
		assert.NoError(t, err)
	})

	t.Run("Delete not existing member in group", func(t *testing.T) {
		user := mustMakeUser(t, repo, mustNewUUIDV4(t), false)
		err := repo.DeleteUserInGroup(group.ID, user.ID)
		// TODO fix
		assert.NoError(t, err)
	})

}

func TestGormRepository_GetGroup(t *testing.T) {
	t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)
	group := mustMakeGroup(t, repo, traQutils.RandAlphabetAndNumberString(20), user.ID)
	mustAddGroupMember(t, repo, group.ID, user.ID)

	t.Run("Get existing group", func(t *testing.T) {
		if group, err := repo.GetGroup(group.ID); assert.NoError(t, err) {
			assert.NotNil(t, group)
			assert.Equal(t, user.ID, group.Members[0].ID)
		}
	})

	t.Run("Get not existing group", func(t *testing.T) {
		_, err := repo.GetGroup(mustNewUUIDV4(t))
		assert.EqualError(t, err, ErrNotFound.Error())
	})
}

func TestGormRepository_GetUserBelongingGroupIDs(t *testing.T) {
	repo, _, _, user := setupGormRepoWithUser(t, common)
	group1 := mustMakeGroup(t, repo, traQutils.RandAlphabetAndNumberString(20), user.ID)
	mustAddGroupMember(t, repo, group1.ID, user.ID)

	group2 := mustMakeGroup(t, repo, traQutils.RandAlphabetAndNumberString(20), user.ID)
	mustAddGroupMember(t, repo, group2.ID, user.ID)

	t.Run("Success", func(t *testing.T) {
		if groupIDs, err := repo.GetUserBelongingGroupIDs(user.ID); assert.NoError(t, err) {
			assert.Len(t, groupIDs, 2)
		}
	})
}
