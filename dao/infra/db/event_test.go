package db

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/knoQ/domain"
)

func Test_createEvent(t *testing.T) {
	r := repositories[common]

	user, _ := saveUser(r.db, mustNewUUIDV4(t), true, true)
	room, _ := createRoom(r.db, writeRoomParams{
		Verified:  false,
		CreatedBy: user.ID,
		WriteRoomParams: domain.WriteRoomParams{
			Place:     "here",
			TimeStart: time.Now(),
			TimeEnd:   time.Now().Add(1 * time.Hour),
		},
	})

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
		events, err := getAllEvents(r.db)
		if assert.NoError(t, err) {
			assert.NotNil(t, events[0].Tags[0].Tag.Name)
		}
	}
}
