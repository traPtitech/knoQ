package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	traQutils "github.com/traPtitech/traQ/utils"
)

func TestGormRepository_CreateEvent(t *testing.T) {
	t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)
	// group := mustMakeGroup(t, repo, traQutils.RandAlphabetAndNumberString(10), user.ID)
	// room := mustMakeRoom(t, repo, traQutils.RandAlphabetAndNumberString(10))

	params := WriteEventParams{
		Name: traQutils.RandAlphabetAndNumberString(20),
		//GroupID:   group.ID,
		//RoomID:    room.ID,
		TimeStart: time.Now(),
		TimeEnd:   time.Now().Add(1 * time.Hour),
		CreatedBy: user.ID,
	}

	if event, err := repo.CreateEvent(params); assert.NoError(t, err) {
		assert.NotNil(t, event)
	}

}
