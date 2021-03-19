package db

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/knoQ/domain"
)

func Test_createEvent(t *testing.T) {
	r, _, _, user, room := setupRepoWithUserRoom(t, common)

	event, err := createEvent(r.db, writeEventParams{
		CreatedBy: user.ID,
		WriteEventParams: domain.WriteEventParams{
			Name:      "first event",
			GroupID:   mustNewUUIDV4(t),
			RoomID:    room.ID,
			TimeStart: time.Now(),
			TimeEnd:   time.Now().Add(1 * time.Hour),
			Admins:    []uuid.UUID{user.ID},
			Tags: []domain.EventTagParams{
				{Name: "go", Locked: true}, {Name: "golang"},
			},
		},
	})

	if assert.NoError(t, err) {
		assert.NotNil(t, event.ID)
		// TODO wip
		events, err := getAllEvents(r.db)
		if assert.NoError(t, err) {
			assert.NotNil(t, events[0].Tags[0].Tag.Name)
		}

		tag, _ := createTag(r.db, "Go")
		err = addEventTag(r.db, event.ID, domain.WriteTagRelationParams{
			ID:     tag.ID,
			Locked: true,
		})
		assert.NoError(t, err)
	}
}
