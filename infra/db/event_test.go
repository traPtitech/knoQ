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

	params := writeEventParams{
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
	})

}
