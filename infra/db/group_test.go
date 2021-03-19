package db

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/knoQ/domain"
)

func Test_createGroup(t *testing.T) {

	r, _, _, user := setupRepoWithUser(t, common)

	group, err := createGroup(r.db, writeGroupParams{
		CreatedBy: user.ID,
		WriteGroupParams: domain.WriteGroupParams{
			Name:    "first group",
			Members: []uuid.UUID{user.ID},
			Admins:  []uuid.UUID{user.ID},
		},
	})

	if assert.NoError(t, err) {
		assert.NotNil(t, group.ID)
	}
}
