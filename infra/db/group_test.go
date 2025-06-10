package db

import (
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/gofrs/uuid"
	"github.com/jinzhu/copier"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

func Test_createGroup(t *testing.T) {
	r, assert, require, user := setupRepoWithUser(t, common)

	params := WriteGroupParams{
		CreatedBy: user.ID,
		WriteGroupParams: domain.WriteGroupParams{
			Name:    "first group",
			Members: []uuid.UUID{user.ID},
			Admins:  []uuid.UUID{user.ID},
		},
	}

	t.Run("create group", func(_ *testing.T) {
		group, err := createGroup(r.db, params)
		require.NoError(err)
		assert.NotNil(group.ID)
	})

	t.Run("create group with invalid members", func(t *testing.T) {
		var p WriteGroupParams
		require.NoError(copier.Copy(&p, &params))
		p.Members = append(p.Members, mustNewUUIDV4(t))
		_, err := createGroup(r.db, p)
		var me *mysql.MySQLError
		require.ErrorAs(err, &me)
		assert.Equal(uint16(1452), me.Number)
		assert.Contains(me.Message, "group_members")
	})

	t.Run("create group with invalid admins", func(t *testing.T) {
		var p WriteGroupParams
		require.NoError(copier.Copy(&p, &params))

		p.Admins = nil
		_, err := createGroup(r.db, p)
		assert.ErrorIs(err, ErrNoAdmins)

		p.Admins = []uuid.UUID{mustNewUUIDV4(t)}
		_, err = createGroup(r.db, p)

		var me *mysql.MySQLError
		require.ErrorAs(err, &me)
		assert.Equal(uint16(1452), me.Number)
		assert.Contains(me.Message, "group_admins")
	})
}

func Test_updateGroup(t *testing.T) {
	r, assert, require, user, group := setupRepoWithUserGroup(t, common)

	params := WriteGroupParams{
		CreatedBy: user.ID,
		WriteGroupParams: domain.WriteGroupParams{
			Name:    "update group",
			Members: []uuid.UUID{user.ID},
			Admins:  []uuid.UUID{user.ID},
		},
	}

	t.Run("update group", func(_ *testing.T) {
		_, err := updateGroup(r.db, group.ID, params)
		require.NoError(err)

		g, err := getGroup(groupFullPreload(r.db), group.ID)
		require.NoError(err)
		assert.Len(g.Members, len(params.Members))
	})

	t.Run("update random groupID", func(t *testing.T) {
		_, err := updateGroup(r.db, mustNewUUIDV4(t), params)
		var me *mysql.MySQLError
		require.ErrorAs(err, &me)
		assert.Equal(uint16(1452), me.Number)
	})
}

func Test_addMemberToGroup(t *testing.T) {
	r, assert, require, _, group := setupRepoWithUserGroup(t, common)

	t.Run("add a member", func(t *testing.T) {
		user := mustMakeUser(t, r, false)
		err := addMemberToGroup(r.db, group.ID, user.ID)
		require.NoError(err)

		g, err := getGroup(r.db.Preload("Members"), group.ID)
		require.NoError(err)
		assert.Len(g.Members, 1)
	})

	t.Run("add a member to random groupID", func(t *testing.T) {
		user := mustMakeUser(t, r, false)
		err := addMemberToGroup(r.db, mustNewUUIDV4(t), user.ID)
		var me *mysql.MySQLError
		require.ErrorAs(err, &me)
		assert.Equal(uint16(1452), me.Number)
	})

	t.Run("add a invalid member", func(t *testing.T) {
		err := addMemberToGroup(r.db, group.ID, mustNewUUIDV4(t))
		var me *mysql.MySQLError
		require.ErrorAs(err, &me)
		assert.Equal(uint16(1452), me.Number)
	})

	t.Run("add a duplicate member", func(t *testing.T) {
		user := mustMakeUser(t, r, false)
		err := addMemberToGroup(r.db, group.ID, user.ID)
		require.NoError(err)

		err = addMemberToGroup(r.db, group.ID, user.ID)
		assert.NoError(err)
	})

	t.Run("add a member who had deleted", func(t *testing.T) {
		user := mustMakeUser(t, r, false)
		err := addMemberToGroup(r.db, group.ID, user.ID)
		require.NoError(err)

		err = deleteMemberOfGroup(r.db, group.ID, user.ID)
		require.NoError(err)

		err = addMemberToGroup(r.db, group.ID, user.ID)
		assert.NoError(err)
	})
}

func Test_deleteGroup(t *testing.T) {
	r, assert, _, _, group := setupRepoWithUserGroup(t, common)

	t.Run("delete group", func(_ *testing.T) {
		err := deleteGroup(r.db, group.ID)
		assert.NoError(err)
	})

	t.Run("delete random groupID", func(t *testing.T) {
		err := deleteGroup(r.db, mustNewUUIDV4(t))
		assert.NoError(err)
	})
}

func Test_deleteMemberOfGroup(t *testing.T) {
	r, assert, require, user, group := setupRepoWithUserGroup(t, common)

	t.Run("delete member", func(_ *testing.T) {
		err := addMemberToGroup(r.db, group.ID, user.ID)
		require.NoError(err)

		err = deleteMemberOfGroup(r.db, group.ID, user.ID)
		require.NoError(err)
		g, err := getGroup(r.db.Preload("Members"), group.ID)
		require.NoError(err)
		assert.Len(g.Members, 0)
	})

	t.Run("delete invalid member", func(t *testing.T) {
		err := deleteMemberOfGroup(r.db.Debug(), group.ID, mustNewUUIDV4(t))
		assert.NoError(err)
	})
}

func Test_getGroup(t *testing.T) {
	r, assert, require, _, group := setupRepoWithUserGroup(t, common)

	t.Run("get group", func(_ *testing.T) {
		g, err := getGroup(r.db, group.ID)
		require.NoError(err)
		assert.Equal(group.Name, g.Name)
	})

	t.Run("get random groupID", func(t *testing.T) {
		_, err := getGroup(r.db, mustNewUUIDV4(t))
		assert.ErrorIs(err, gorm.ErrRecordNotFound)
	})
}
