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
	userID, _ := uuid.NewV4()
	user := mustMakeUser(t, repo, userID, false)

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
