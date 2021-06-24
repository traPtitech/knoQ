package redis

import (
	"context"

	"github.com/go-redis/cache/v8"
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
)

func (repo *RedisRepository) SetGroup(group *domain.Group, info *domain.ConInfo) error {
	repo.setValidUser(info.ReqUserID)

	ctx := context.TODO()
	return repo.groupsCache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   group.ID.String(),
		Value: group,
		TTL:   repo.groupsCacheTime,
	})
}

func (repo *RedisRepository) GetGroup(groupID uuid.UUID, info *domain.ConInfo) (*domain.Group, error) {
	if !repo.isValidUser(info.ReqUserID) {
		return nil, ErrValidationExpired
	}

	ctx := context.TODO()
	var group domain.Group
	err := repo.groupsCache.Get(ctx, groupID.String(), &group)
	return &group, err
}

func (repo *RedisRepository) SetGroups(groups []*domain.Group, info *domain.ConInfo) error {
	repo.setValidUser(info.ReqUserID)

	ctx := context.TODO()
	return repo.groupsCache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   "groups",
		Value: groups,
		TTL:   repo.groupsCacheTime,
	})
}

func (repo *RedisRepository) GetGroups(info *domain.ConInfo) ([]*domain.Group, error) {
	if !repo.isValidUser(info.ReqUserID) {
		return nil, ErrValidationExpired
	}

	ctx := context.TODO()
	var groups []*domain.Group
	err := repo.groupsCache.Get(ctx, "groups", &groups)
	return groups, err
}
