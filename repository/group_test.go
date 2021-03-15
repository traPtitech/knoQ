package repository

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	traQrandom "github.com/traPtitech/traQ/utils/random"
)

func TestGormRepository_CreateGroup(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupGormRepo(t, common)
	user := mustMakeUserMeta(t, repo, false)

	params := WriteGroupParams{
		Name:      traQrandom.AlphaNumeric(20),
		Members:   []uuid.UUID{user.ID, mustNewUUIDV4(t)},
		Admins:    []uuid.UUID{user.ID},
		CreatedBy: user.ID,
	}

	if g, err := repo.CreateGroup(params); assert.NoError(t, err) {
		if group, err := repo.GetGroup(g.ID); assert.NoError(t, err) {
			assert.NotNil(t, group)
			assert.Equal(t, 2, len(group.Members))
			assert.Equal(t, 1, len(group.Admins))
		}
	}
}

func TestGormRepository_UpdateGroup(t *testing.T) {
	//t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)
	group := mustMakeGroup(t, repo, traQrandom.AlphaNumeric(20), user.ID)

	params := WriteGroupParams{
		Name:      group.Name,
		Members:   []uuid.UUID{user.ID, mustNewUUIDV4(t)},
		Admins:    []uuid.UUID{user.ID},
		CreatedBy: user.ID,
	}

	if g, err := repo.UpdateGroup(group.ID, params); assert.NoError(t, err) {
		if group, err := repo.GetGroup(g.ID); assert.NoError(t, err) {
			assert.NotNil(t, group)
			assert.Equal(t, 2, len(group.Members))
			assert.Equal(t, 1, len(group.Admins))
		}
	}
}

func TestGormRepository_AddUserToGroup(t *testing.T) {
	//t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)
	group := mustMakeGroup(t, repo, traQrandom.AlphaNumeric(20), user.ID)

	t.Run("Add existing user", func(t *testing.T) {
		t.Parallel()
		err := repo.AddUserToGroup(group.ID, user.ID)
		assert.NoError(t, err)
	})

	t.Run("Add not existing user", func(t *testing.T) {
		t.Parallel()
		err := repo.AddUserToGroup(group.ID, mustNewUUIDV4(t))
		assert.NoError(t, err)
	})

	t.Run("Add already exists user", func(t *testing.T) {
		t.Parallel()
		user := mustMakeUserMeta(t, repo, false)
		err := repo.AddUserToGroup(group.ID, user.ID)
		assert.NoError(t, err)
		err = repo.AddUserToGroup(group.ID, user.ID)
		assert.EqualError(t, err, ErrAlreadyExists.Error())
	})
}

func TestGormRepository_DeleteGroup(t *testing.T) {
	t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)
	group := mustMakeGroup(t, repo, traQrandom.AlphaNumeric(20), user.ID)

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
	group := mustMakeGroup(t, repo, traQrandom.AlphaNumeric(20), user.ID)
	mustAddGroupMember(t, repo, group.ID, user.ID)

	t.Run("Delete existing member in group", func(t *testing.T) {
		t.Parallel()
		err := repo.DeleteUserInGroup(group.ID, user.ID)
		assert.NoError(t, err)
	})

	t.Run("Delete not existing member in group", func(t *testing.T) {
		user := mustMakeUserMeta(t, repo, false)
		err := repo.DeleteUserInGroup(group.ID, user.ID)
		assert.EqualError(t, err, ErrNotFound.Error())
	})

}

func TestGormRepository_GetGroup(t *testing.T) {
	t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)
	group := mustMakeGroup(t, repo, traQrandom.AlphaNumeric(20), user.ID)
	mustAddGroupMember(t, repo, group.ID, user.ID)

	t.Run("Get existing group", func(t *testing.T) {
		if group, err := repo.GetGroup(group.ID); assert.NoError(t, err) {
			assert.NotNil(t, group)
			assert.Equal(t, user.ID, group.Members[0].UserID)
		}
	})

	t.Run("Get not existing group", func(t *testing.T) {
		_, err := repo.GetGroup(mustNewUUIDV4(t))
		assert.EqualError(t, err, ErrNotFound.Error())
	})
}

func TestGormRepository_GetUserBelongingGroupIDs(t *testing.T) {
	repo, _, _, user := setupGormRepoWithUser(t, common)
	group1 := mustMakeGroup(t, repo, traQrandom.AlphaNumeric(20), user.ID)
	mustAddGroupMember(t, repo, group1.ID, user.ID)

	group2 := mustMakeGroup(t, repo, traQrandom.AlphaNumeric(20), user.ID)
	mustAddGroupMember(t, repo, group2.ID, user.ID)

	t.Run("success", func(t *testing.T) {
		if groupIDs, err := repo.GetUserBelongingGroupIDs(user.ID); assert.NoError(t, err) {
			assert.Len(t, groupIDs, 2)
		}
	})
}
func TestTraQRepository_CreateGroup(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupTraQRepo(t, TraQv3)
	user := mustMakeUserBody(t, repo, traQrandom.AlphaNumeric(20), traQrandom.AlphaNumeric(20))

	params := WriteGroupParams{
		Name:      traQrandom.AlphaNumeric(20),
		Members:   []uuid.UUID{user.ID},
		CreatedBy: user.ID,
	}

	if group, err := repo.CreateGroup(params); assert.NoError(t, err) {
		assert.NotNil(t, group)
		assert.Equal(t, params.Members[0], group.Members[0].UserID)
		assert.Equal(t, 1, len(group.Members))
	}

}

func TestTraQRepository_GetUserBelongingGroupIDs(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupTraQRepo(t, TraQv1)
	user := mustMakeUserBody(t, repo, traQrandom.AlphaNumeric(10), traQrandom.AlphaNumeric(10))

	t.Run("success", func(t *testing.T) {
		if groupIDs, err := repo.GetUserBelongingGroupIDs(user.ID); assert.NoError(t, err) {
			assert.NotNil(t, groupIDs)
		}
	})
}

func TestTraQRepository_GetGroup(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupTraQRepo(t, TraQv3)
	user := mustMakeUserBody(t, repo, traQrandom.AlphaNumeric(10), traQrandom.AlphaNumeric(10))
	group := mustMakeGroup(t, repo, traQrandom.AlphaNumeric(10), user.ID)

	t.Run("success", func(t *testing.T) {
		if group, err := repo.GetGroup(group.ID); assert.NoError(t, err) {
			assert.NotNil(t, group)
		}
	})
}

func TestTraPGroupRepository_GetGroup(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupTraPGroupRepo(t, TraQv3)

	t.Run("success", func(t *testing.T) {
		groupID, err := uuid.FromString("11111111-1111-1111-1111-111111111111")
		require.NoError(t, err)
		if group, err := repo.GetGroup(groupID); assert.NoError(t, err) {
			assert.NotNil(t, group)
		}
	})

	t.Run("Get not existing groupID", func(t *testing.T) {
		_, err := repo.GetGroup(mustNewUUIDV4(t))
		assert.EqualError(t, err, ErrNotFound.Error())
	})
}
