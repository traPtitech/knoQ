package repository

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
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

func TestGormRepository_UpdateEvent(t *testing.T) {
	repo, _, _, user := setupGormRepoWithUser(t, common)
	e, _, _ := mustMakeEvent(t, repo, traQutils.RandAlphabetAndNumberString(10), user.ID)

	params := WriteEventParams{
		Name:      e.Name,
		GroupID:   e.GroupID,
		RoomID:    e.RoomID,
		TimeStart: time.Now(),
		TimeEnd:   time.Now().Add(10 * time.Minute),
		CreatedBy: e.CreatedBy,
	}

	if event, err := repo.UpdateEvent(e.ID, params); assert.NoError(t, err) {
		assert.Equal(t, e.ID, event.ID)
		assert.Equal(t, params.TimeEnd, event.TimeEnd)
	}
}

func TestGormRepository_GetEventsByGroupIDs(t *testing.T) {
	repo, _, _, user := setupGormRepoWithUser(t, common)
	_, g1, _ := mustMakeEvent(t, repo, traQutils.RandAlphabetAndNumberString(10), user.ID)
	_, g2, _ := mustMakeEvent(t, repo, traQutils.RandAlphabetAndNumberString(10), user.ID)

	if events, err := repo.GetEventsByGroupIDs([]uuid.UUID{g1.ID, g2.ID}); assert.NoError(t, err) {
		assert.Equal(t, 2, len(events))
	}
}
