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
		TTL:   repo.usersCacheTime,
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

func (repo *RedisRepository) DeleteUser(userID uuid.UUID, info *domain.ConInfo) (*domain.User, error) {
	ctx := context.TODO()
	var user domain.User
	err := repo.usersCache.Delete(ctx, userID.String())
	return &user, err
}

func (repo *RedisRepository) SetUsers(users []*domain.User, info *domain.ConInfo) error {
	repo.setValidUser(info.ReqUserID)

	ctx := context.TODO()
	return repo.usersCache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   "users",
		Value: users,
		TTL:   repo.usersCacheTime,
	})
}

func (repo *RedisRepository) GetUsers(info *domain.ConInfo) ([]*domain.User, error) {
	if !repo.isValidUser(info.ReqUserID) {
		return nil, ErrValidationExpired
	}

	ctx := context.TODO()
	var users []*domain.User
	err := repo.usersCache.Get(ctx, "users", &users)
	return users, err
}

func (repo *RedisRepository) DeleteUsers(info *domain.ConInfo) ([]*domain.User, error) {
	ctx := context.TODO()
	var users []*domain.User
	err := repo.usersCache.Delete(ctx, "users")
	return users, err
}
