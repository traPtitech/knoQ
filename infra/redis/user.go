package redis

import (
	"context"

	"github.com/go-redis/cache/v8"
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
)

func (repo *RedisRepository) SetUser(user *domain.User, info *domain.ConInfo) error {
	repo.setValidUser(info.ReqUserID)

	ctx := context.TODO()
	return repo.usersCache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   user.ID.String(),
		Value: user,
	})
}

func (repo *RedisRepository) GetUser(userID uuid.UUID, info *domain.ConInfo) (*domain.User, error) {
	if !repo.isValidUser(info.ReqUserID) {
		return nil, ErrValidationExpired
	}

	ctx := context.TODO()
	var user domain.User
	err := repo.usersCache.Get(ctx, userID.String(), &user)
	return &user, err
}
