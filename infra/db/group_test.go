package db

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/copier"
	"github.com/traPtitech/knoQ/domain"
)

func Test_createGroup(t *testing.T) {
	r, assert, require, user := setupRepoWithUser(t, common)

	params := writeGroupParams{
		CreatedBy: user.ID,
		WriteGroupParams: domain.WriteGroupParams{
			Name:    "first group",
			Members: []uuid.UUID{user.ID},
			Admins:  []uuid.UUID{user.ID},
		},
	}

	t.Run("create group", func(t *testing.T) {
		group, err := createGroup(r.db, params)
		require.NoError(err)
		assert.NotNil(group.ID)
	})

	t.Run("create group with invalid members", func(t *testing.T) {
		var p writeGroupParams
		require.NoError(copier.Copy(&p, &params))
		p.Members = append(p.Members, mustNewUUIDV4(t))
		_, err := createGroup(r.db, params)
		require.NoError(err)
	})
}
