package db

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/knoQ/domain"
)

func Test_createEvent(t *testing.T) {
	r, _, _, user, room := setupRepoWithUserRoom(t, common)

	params := WriteEventParams{
		CreatedBy: user.ID,
		WriteEventParams: domain.WriteEventParams{
			Name:          "first event",
			GroupID:       mustNewUUIDV4(t),
			RoomID:        room.ID,
			TimeStart:     time.Now(),
			TimeEnd:       time.Now().Add(1 * time.Minute),
			AllowTogether: true,
			Admins:        []uuid.UUID{user.ID},
			Tags: []domain.EventTagParams{
				{Name: "go", Locked: true}, {Name: "golang"},
			},
		},
	}

	t.Run("create event", func(t *testing.T) {
		event, err := createEvent(r.db, params)
		require.NoError(t, err)
		assert.NotNil(t, event.ID)

		// tags
		e, err := getEvent(r.db.Preload("Tags").Preload("Tags.Tag"), event.ID)
		require.NoError(t, err)
		assert.NotNil(t, e.Tags[0].Tag.Name)
	})

	t.Run("create event with exsiting tags", func(t *testing.T) {
		_, err := createTag(r.db, "Go")
		require.NoError(t, err)

		params.Tags = append(params.Tags, domain.EventTagParams{Name: "Go"})
		_, err = createEvent(r.db, params)
		require.NoError(t, err)
	})

	t.Run("wrong time", func(t *testing.T) {
		params.TimeStart = time.Now().Add(10 * time.Minute)
		_, err := createEvent(r.db, params)
		assert.ErrorIs(t, err, ErrTimeConsistency)
	})

}

func Test_updateEvent(t *testing.T) {
	r, _, _, user, _, room, event := setupRepoWithUserGroupRoomEvent(t, common)

	params := WriteEventParams{
		CreatedBy: user.ID,
		WriteEventParams: domain.WriteEventParams{
			Name:          "update event",
			GroupID:       mustNewUUIDV4(t),
			RoomID:        room.ID,
			TimeStart:     time.Now(),
			TimeEnd:       time.Now().Add(1 * time.Minute),
			AllowTogether: true,
			Admins:        []uuid.UUID{user.ID},
			Tags: []domain.EventTagParams{
				{Name: "go", Locked: true}, {Name: "golang2"},
			},
		},
	}

	t.Run("update event", func(t *testing.T) {
		_, err := updateEvent(r.db, event.ID, params)
		require.NoError(t, err)

		e, err := getEvent(r.db, event.ID)
		require.NoError(t, err)

		assert.Equal(t, len(params.Tags), len(e.Tags))
	})
}

func Test_addEventTag(t *testing.T) {
	r, _, _, _, _, _, event := setupRepoWithUserGroupRoomEvent(t, common)

	t.Run("add tag", func(t *testing.T) {
		err := addEventTag(r.db.Debug(), event.ID, domain.EventTagParams{
			Name: "foo",
		})
		require.NoError(t, err)
	})
}
