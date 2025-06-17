package db

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/copier"
	"github.com/traPtitech/knoQ/domain"
)

func Test_createRoom(t *testing.T) {
	r, assert, require, user := setupRepoWithUser(t, common)

	params := CreateRoomParams{
		CreatedBy: user.ID,
		WriteRoomParams: domain.WriteRoomParams{
			Place:     "create room",
			TimeStart: time.Now(),
			TimeEnd:   time.Now().Add(1 * time.Minute),
			Admins:    []uuid.UUID{user.ID},
		},
	}

	t.Run("create room", func(_ *testing.T) {
		room, err := createRoom(r.db, params)
		require.NoError(err)
		assert.NotNil(room.ID)
	})

	t.Run("wrong time", func(_ *testing.T) {
		var p CreateRoomParams
		require.NoError(copier.Copy(&p, &params))

		p.TimeStart = time.Now().Add(10 * time.Minute)
		_, err := createRoom(r.db, p)
		assert.ErrorIs(err, ErrTimeConsistency)
	})
}

func Test_updateRoom(t *testing.T) {
	r, assert, require, user, room := setupRepoWithUserRoom(t, common)

	params := UpdateRoomParams{
		CreatedBy: user.ID,
		WriteRoomParams: domain.WriteRoomParams{
			Place:     "update room",
			TimeStart: time.Now(),
			TimeEnd:   time.Now().Add(1 * time.Minute),
			Admins:    []uuid.UUID{user.ID},
		},
	}

	t.Run("update room", func(_ *testing.T) {
		_, err := updateRoom(r.db, room.ID, params)
		require.NoError(err)

		ro, err := getRoom(roomFullPreload(r.db), room.ID)
		require.NoError(err)

		assert.Equal(params.Place, ro.Name)
	})

	t.Run("update random roomID", func(t *testing.T) {
		_, err := updateRoom(r.db, mustNewUUIDV4(t), params)
		assert.Error(err)
		// var me *mysql.MySQLError
		// require.ErrorAs(err, &me)
		// assert.Equal(uint16(1452), me.Number)
	})
}
