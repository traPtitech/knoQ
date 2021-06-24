package redis

import (
	"context"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
)

var (
	usersCacheTime  = 1 * time.Minute
	groupsCacheTime = 15 * time.Second
	validCacheTime  = 5 * time.Minute
)

type RedisRepository struct {
	usersCache  *cache.Cache
	groupsCache *cache.Cache

	// キャッシュを許可するユーザーキャッシュ
	validCache *cache.Cache
}

func Setup(host, port string) *RedisRepository {
	addrs := map[string]string{
		host: ":" + port,
	}

	rings := make([]*redis.Ring, 16)
	for i := range rings {
		rings[i] = redis.NewRing(&redis.RingOptions{
			Addrs: addrs,
			DB:    i,
		})
	}

	repo := new(RedisRepository)
	repo.usersCache = cache.New(&cache.Options{
		Redis: rings[0],
	})
	repo.groupsCache = cache.New(&cache.Options{
		Redis: rings[1],
	})
	repo.validCache = cache.New(&cache.Options{
		Redis: rings[2],
	})

	return repo
}

func (repo *RedisRepository) setValidUser(userID uuid.UUID) error {
	ctx := context.TODO()
	return repo.validCache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   userID.String(),
		Value: true,
		TTL:   validCacheTime,
	})
}

func (repo *RedisRepository) isValidUser(userID uuid.UUID) bool {
	ctx := context.TODO()
	return repo.validCache.Exists(ctx, userID.String())
}
