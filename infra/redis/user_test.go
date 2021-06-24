package redis

import (
	"testing"

	"github.com/go-redis/cache/v8"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/traQ/utils/random"
)

func TestRedisRepository_GetUser(t *testing.T) {
	r, assert, require := setupRepo(t)

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
}
