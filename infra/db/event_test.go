package db

import (
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gofrs/uuid"
	"github.com/jinzhu/copier"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/utils/random"
	"gorm.io/gorm"
)

func Test_createEvent(t *testing.T) {
	r, assert, require, user, room := setupRepoWithUserRoom(t, common)

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
		require.NoError(err)
		assert.NotNil(event.ID)

		// tags
		e, err := getEvent(r.db.Preload("Tags").Preload("Tags.Tag"), event.ID)
		require.NoError(err)
		assert.NotNil(e.Tags[0].Tag.Name)
	})

	t.Run("create event with exsiting tags", func(t *testing.T) {
		_, err := createOrGetTag(r.db, "Go")
		require.NoError(err)

		var p WriteEventParams
		require.NoError(copier.Copy(&p, &params))

		p.Tags = append(p.Tags, domain.EventTagParams{Name: "Go"})
		_, err = createEvent(r.db, p)
		require.NoError(err)
	})

	t.Run("wrong time", func(t *testing.T) {
		var p WriteEventParams
		require.NoError(copier.Copy(&p, &params))

		p.TimeStart = time.Now().Add(10 * time.Minute)
		_, err := createEvent(r.db, p)
		assert.ErrorIs(err, ErrTimeConsistency)
	})

	t.Run("wrong room time", func(t *testing.T) {
		var p WriteEventParams
		require.NoError(copier.Copy(&p, &params))

		p.AllowTogether = false
		_, err := createEvent(r.db, p)
		assert.ErrorIs(err, ErrTimeConsistency)
	})

	t.Run("create event with place", func(t *testing.T) {
		var p WriteEventParams
		require.NoError(copier.Copy(&p, &params))

		p.RoomID = uuid.Nil
		p.Place = "instant room"
		event, err := createEvent(r.db.Debug(), p)
		require.NoError(err)

		e, err := getEvent(eventFullPreload(r.db), event.ID)
		require.NoError(err)
		assert.NotEqual(uuid.Nil, e.RoomID)
		assert.Equal(p.Place, e.Room.Place)
	})
}

func Test_updateEvent(t *testing.T) {
	r, assert, require, user, _, room, event := setupRepoWithUserGroupRoomEvent(t, common)

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
		require.NoError(err)

		e, err := getEvent(eventFullPreload(r.db), event.ID)
		require.NoError(err)

		assert.Equal(len(params.Tags), len(e.Tags))
	})

	t.Run("update random eventID", func(t *testing.T) {
		_, err := updateEvent(r.db, mustNewUUIDV4(t), params)
		var me *mysql.MySQLError
		require.ErrorAs(err, &me)
		assert.Equal(uint16(1452), me.Number)
	})
}

func Test_addEventTag(t *testing.T) {
	r, assert, require, _, _, _, event := setupRepoWithUserGroupRoomEvent(t, common)

	t.Run("add tag", func(t *testing.T) {
		err := addEventTag(r.db, event.ID, domain.EventTagParams{
			Name: "foo",
		})
		require.NoError(err)
	})

	t.Run("add tag in random eventID", func(t *testing.T) {
		err := addEventTag(r.db, mustNewUUIDV4(t), domain.EventTagParams{
			Name: "foo",
		})
		var me *mysql.MySQLError
		require.ErrorAs(err, &me)
		assert.Equal(uint16(1452), me.Number)
	})
}

func Test_deleteEvent(t *testing.T) {
	r, assert, require, _, _, _, event := setupRepoWithUserGroupRoomEvent(t, common)

	t.Run("delete event", func(t *testing.T) {
		err := deleteEvent(r.db, event.ID)
		require.NoError(err)

		_, err = getEvent(r.db, event.ID)
		assert.ErrorIs(err, gorm.ErrRecordNotFound)
	})

	t.Run("delete random eventID", func(t *testing.T) {
		err := deleteEvent(r.db, mustNewUUIDV4(t))
		assert.NoError(err)
	})
}

func Test_deleteEventTag(t *testing.T) {
	r, assert, require, _, _, _, event := setupRepoWithUserGroupRoomEvent(t, common)

	t.Run("delete eventTag", func(t *testing.T) {
		err := deleteEventTag(r.db, event.ID, "gin", false)
		require.NoError(err)

		e, err := getEvent(r.db.Preload("Tags").Preload("Tags.Tag"), event.ID)
		require.NoError(err)
		assert.Empty(e.Tags)
	})

	t.Run("delete locked tag", func(t *testing.T) {
		err := addEventTag(r.db, event.ID, domain.EventTagParams{
			Name: "LOCK", Locked: true,
		})
		require.NoError(err)

		err = deleteEventTag(r.db, event.ID, "LOCK", false)
		assert.NoError(err)
		e, err := getEvent(r.db.Preload("Tags").Preload("Tags.Tag"), event.ID)
		require.NoError(err)
		assert.True(containsEventTag(e.Tags, "LOCK"))
	})

	t.Run("delete tag in random eventID", func(t *testing.T) {
		err := addEventTag(r.db, event.ID, domain.EventTagParams{
			Name: "gin2",
		})
		require.NoError(err)
		err = deleteEventTag(r.db, mustNewUUIDV4(t), "gin2", false)
		assert.NoError(err)

		err = deleteEventTag(r.db, mustNewUUIDV4(t), random.AlphaNumeric(8, false), false)
		assert.ErrorIs(err, gorm.ErrRecordNotFound)
	})

	t.Run("delete non-tag", func(t *testing.T) {
		err := deleteEventTag(r.db, event.ID, random.AlphaNumeric(8, false), false)
		assert.ErrorIs(err, gorm.ErrRecordNotFound)
	})
}

func Test_getEvent(t *testing.T) {
	r, assert, require, _, _, _, event := setupRepoWithUserGroupRoomEvent(t, common)

	t.Run("get Event", func(t *testing.T) {
		e, err := getEvent(r.db, event.ID)
		require.NoError(err)
		assert.Equal(event.Name, e.Name)
	})

	t.Run("get random eventID", func(t *testing.T) {
		_, err := getEvent(r.db, mustNewUUIDV4(t))
		assert.ErrorIs(err, gorm.ErrRecordNotFound)
	})
}

func containsEventTag(tags []EventTag, tagName string) (exist bool) {
	exist = false
	for _, tag := range tags {
		if tag.Tag.Name == tagName {
			exist = true
			return
		}
	}
	return
}
