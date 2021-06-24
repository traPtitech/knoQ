package redis

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	common = "common"
	onesec = "onesec"
)

var (
	repositories = map[string]*RedisRepository{}
)

func TestMain(m *testing.M) {
	const (
		host = "localhost"
		port = "6379"
	)
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"server": host + ":" + port,
		},
	})
	ctx := context.Background()
	// delete all
	ring.FlushAll(ctx)

	repositories[common] = Setup(host, port)

	r := Setup(host, port)
	r.usersCacheTime = 1 * time.Second
	r.groupsCacheTime = 1 * time.Second
	r.validCacheTime = 1 * time.Second
	repositories[onesec] = r

	m.Run()
}

func assertAndRequire(t *testing.T) (*assert.Assertions, *require.Assertions) {
	return assert.New(t), require.New(t)
}

func mustNewUUIDV4(t *testing.T) uuid.UUID {
	id, err := uuid.NewV4()
	require.NoError(t, err)
	return id
}

func setupRepo(t *testing.T, repo string) (*RedisRepository, *assert.Assertions, *require.Assertions) {
	t.Helper()
	r, ok := repositories[repo]
	if !ok {
		t.FailNow()
	}
	assert, require := assertAndRequire(t)
	return r, assert, require
}
