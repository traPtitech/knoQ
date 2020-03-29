package service

import (
	repo "room/repository"

	"github.com/gofrs/uuid"
)

// Dao DataAccess Object
type Dao struct {
	Repo                      repo.Repository
	InitExternalUserGroupRepo func(token string, ver repo.TraQVersion) interface {
		repo.GroupRepository
		repo.UserRepository
	}
	ExternalRoomRepo repo.RoomRepository
}

func (d Dao) GetUserBelongingGroupIDs(token string, userID uuid.UUID) ([]uuid.UUID, error) {
	groupIDs, err := d.Repo.GetUserBelongingGroupIDs(userID)
	if err != nil {
		return nil, err
	}

	UserGroupRepo := d.InitExternalUserGroupRepo(token, repo.V3)
	externalGroupIDs, err := UserGroupRepo.GetUserBelongingGroupIDs(userID)
	if err != nil {
		return nil, err
	}
	groupIDs = append(groupIDs, externalGroupIDs...)
	return groupIDs, nil
}
