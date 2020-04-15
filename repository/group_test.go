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
		assert.Equal(t, 2, len(group.Members))
	}
}

func TestGormRepository_UpdateGroup(t *testing.T) {
	//t.Parallel()
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
		assert.Equal(t, 2, len(group.Members))
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
		assert.NoError(t, err)
	})

	t.Run("Add already exists user", func(t *testing.T) {
		t.Parallel()
		user := mustMakeUser(t, repo, mustNewUUIDV4(t), false)
		err := repo.AddUserToGroup(group.ID, user.ID)
		assert.NoError(t, err)
		err = repo.AddUserToGroup(group.ID, user.ID)
		assert.EqualError(t, err, ErrAlreadyExists.Error())
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
		assert.EqualError(t, err, ErrNotFound.Error())
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

func TestTraQRepository_GetUserBelongingGroupIDs(t *testing.T) {
	t.Parallel()
	//repo, _, _ := setupTraQRepo(t, TraQv1)
	//userID, _ := uuid.FromString(os.Getenv("TRAQ_USERID"))

	t.Run("Success", func(t *testing.T) {
		// TODO fix
		//if groupIDs, err := repo.GetUserBelongingGroupIDs(userID); assert.NoError(t, err) {
		//assert.NotNil(t, groupIDs)
		//}
	})
}

func TestTraQRepository_GetGroup(t *testing.T) {
	t.Parallel()
	//repo, _, _ := setupTraQRepo(t, TraQv3)
	//groupID, _ := uuid.FromString(os.Getenv("TRAQ_GROUPID"))

	t.Run("Success", func(t *testing.T) {
		// TODO fix
		//if group, err := repo.GetGroup(groupID); assert.NoError(t, err) {
		//assert.NotNil(t, group)
		//}
	})

}
