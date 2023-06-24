package db

import (
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/utils/random"
)

func Test_createRoom(t *testing.T) {
	r, assert, require, user := setupRepoWithUser(t, common)

	newParams := func() CreateRoomParams {
		return CreateRoomParams{
			CreatedBy: user.ID,
			Verified:  false,
			WriteRoomParams: domain.WriteRoomParams{
				Place:     "create room_" + random.AlphaNumeric(10, false),
				TimeStart: time.Now(),
				TimeEnd:   time.Now().Add(1 * time.Minute),
				Admins:    []uuid.UUID{user.ID},
			},
		}
	}

	t.Run("create room", func(t *testing.T) {
		room, err := createRoom(r.db, newParams())
		require.NoError(err)
		assert.NotNil(room.ID)
	})

	t.Run("wrong time", func(t *testing.T) {
		p := newParams()
		p.TimeStart, p.TimeEnd = p.TimeEnd, p.TimeStart
		_, err := createRoom(r.db, p)
		assert.ErrorIs(err, ErrTimeConsistency)
	})

	t.Run("cannot create room with same place, time", func(t *testing.T) {
		p := newParams()
		_, err := createRoom(r.db, p)
		require.NoError(err)

		_, err = createRoom(r.db, p)
		var me *mysql.MySQLError
		require.ErrorAs(err, &me)
		assert.Equal(uint16(1062), me.Number)
	})
}

func Test_updateRoom(t *testing.T) {
	r, assert, require, user, room := setupRepoWithUserRoom(t, common)

	newParams := func() UpdateRoomParams {
		return UpdateRoomParams{
			CreatedBy: user.ID,
			WriteRoomParams: domain.WriteRoomParams{
				Place:     "update room_" + random.AlphaNumeric(10, false),
				TimeStart: time.Now(),
				TimeEnd:   time.Now().Add(1 * time.Minute),
				Admins:    []uuid.UUID{user.ID},
			},
		}
	}

	t.Run("update room", func(t *testing.T) {
		p := newParams()
		_, err := updateRoom(r.db, room.ID, p)
		require.NoError(err)

		ro, err := getRoom(roomFullPreload(r.db), room.ID)
		require.NoError(err)

		assert.Equal(p.Place, ro.Place)
	})

	t.Run("update room with verified", func(t *testing.T) {
		_p := newParams()
		p := CreateRoomParams{
			WriteRoomParams: _p.WriteRoomParams,
			Verified:        true,
			CreatedBy:       _p.CreatedBy,
		}
		ro, err := createRoom(r.db, p)
		require.NoError(err)

		_, err = updateRoom(r.db, ro.ID, newParams())
		require.NoError(err)

		roo, err := getRoom(r.db, ro.ID)
		require.NoError(err)
		assert.Equal(true, roo.Verified)
	})

	t.Run("update random roomID", func(t *testing.T) {
		_, err := updateRoom(r.db, mustNewUUIDV4(t), newParams())
		var me *mysql.MySQLError
		require.ErrorAs(err, &me)
		assert.Equal(uint16(1452), me.Number)
	})
}
