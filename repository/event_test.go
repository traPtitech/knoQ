package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	traQutils "github.com/traPtitech/traQ/utils"
)

func TestGormRepository_CreateEvent(t *testing.T) {
	t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)
	group := mustMakeGroup(t, repo, traQutils.RandAlphabetAndNumberString(10), user.ID)
	room := mustMakeRoom(t, repo, traQutils.RandAlphabetAndNumberString(10))

	params := WriteEventParams{
		Name:      traQutils.RandAlphabetAndNumberString(20),
		GroupID:   group.ID,
		RoomID:    room.ID,
		TimeStart: time.Now(),
		TimeEnd:   time.Now().Add(1 * time.Minute),
		CreatedBy: user.ID,
	}

	if event, err := repo.CreateEvent(params); assert.NoError(t, err) {
		assert.NotNil(t, event)
	}

}

func TestGormRepository_DeleteTagInEvent(t *testing.T) {
	t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)
	event, _, _ := mustMakeEvent(t, repo, traQutils.RandAlphabetAndNumberString(10), user.ID)

	t.Run("delete unlocked tag when deleteLocked == false", func(t *testing.T) {
		t.Parallel()
		tag := mustMakeTag(t, repo, traQutils.RandAlphabetAndNumberString(10))
		err := repo.AddTagToEvent(event.ID, tag.ID, false)
		require.NoError(t, err)

		err = repo.DeleteTagInEvent(event.ID, tag.ID, false)
		assert.NoError(t, err)

	})

	t.Run("delete locked tag when deleteLocked == false", func(t *testing.T) {
		t.Parallel()
		tag := mustMakeTag(t, repo, traQutils.RandAlphabetAndNumberString(10))
		err := repo.AddTagToEvent(event.ID, tag.ID, true)
		require.NoError(t, err)

		err = repo.DeleteTagInEvent(event.ID, tag.ID, false)
		assert.Error(t, err)
	})

	t.Run("delete locked tag when deleteLocked == true", func(t *testing.T) {
		t.Parallel()
		tag := mustMakeTag(t, repo, traQutils.RandAlphabetAndNumberString(10))
		err := repo.AddTagToEvent(event.ID, tag.ID, true)
		require.NoError(t, err)

		err = repo.DeleteTagInEvent(event.ID, tag.ID, true)
		assert.NoError(t, err)
	})

}
