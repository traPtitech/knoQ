package redis

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/traQ/utils/random"
)

func TestRedisRepository_GetUser(t *testing.T) {
	t.Parallel()
	r, assert, require := setupRepo(t, common)

	user := &domain.User{
		ID:   mustNewUUIDV4(t),
		Name: random.AlphaNumeric(10),
	}
	i := &domain.ConInfo{
		ReqUserID: user.ID,
	}
	err := r.SetUser(user, i)
	require.NoError(err)

	t.Run("get a user", func(t *testing.T) {
		u, err := r.GetUser(user.ID, i)
		require.NoError(err)
		assert.Equal(user.Name, u.Name)
	})

	t.Run("get a random user", func(t *testing.T) {
		_, err := r.GetUser(mustNewUUIDV4(t), i)
		assert.ErrorIs(err, cache.ErrCacheMiss)
	})

	t.Run("bad ConInfo", func(t *testing.T) {
		_, err := r.GetUser(user.ID, &domain.ConInfo{
			ReqUserID: mustNewUUIDV4(t),
		})
		assert.ErrorIs(err, ErrValidationExpired)
	})

	// 1sec test
	r, assert, require = setupRepo(t, onesec)

	user = &domain.User{
		ID:   mustNewUUIDV4(t),
		Name: random.AlphaNumeric(10),
	}
	i = &domain.ConInfo{
		ReqUserID: user.ID,
	}
	err = r.SetUser(user, i)
	require.NoError(err)

	t.Run("wait 2 sec", func(t *testing.T) {
		t.Parallel()
		u, err := r.GetUser(user.ID, i)
		require.NoError(err)
		assert.Equal(user.Name, u.Name)

		time.Sleep(2 * time.Second)
		_, err = r.GetUser(user.ID, i)
		assert.ErrorIs(err, ErrValidationExpired)
	})
}

func TestRedisRepository_GetUsers(t *testing.T) {
	r, assert, require := setupRepo(t, common)

	users := []*domain.User{
		{
			ID:   mustNewUUIDV4(t),
			Name: random.AlphaNumeric(10),
		},
		{
			ID:   mustNewUUIDV4(t),
			Name: random.AlphaNumeric(10),
		},
	}
	i := &domain.ConInfo{
		ReqUserID: users[0].ID,
	}
	err := r.SetUsers(users, i)
	require.NoError(err)

	t.Run("get a user", func(t *testing.T) {
		u, err := r.GetUsers(i)
		require.NoError(err)

		assert.True(reflect.DeepEqual(users, u))
	})
}
