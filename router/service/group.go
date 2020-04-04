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

func (d Dao) GetGroup(token string, groupID uuid.UUID) (*GroupRes, error) {
	group, _ := d.Repo.GetGroup(groupID)
	if group == nil {
		UserGroupRepo := d.InitExternalUserGroupRepo(token, repo.V3)
		group, err := UserGroupRepo.GetGroup(groupID)
		if err != nil {
			return nil, err
		}
		return FormatGroupRes(group, true), nil
	}
	return FormatGroupRes(group, false), nil
}

func (d Dao) GetUserBelongingGroupIDs(token string, userID uuid.UUID) ([]uuid.UUID, error) {
	groupIDs, err := d.Repo.GetUserBelongingGroupIDs(userID)
	if err != nil {
		return nil, err
	}

	UserGroupRepo := d.InitExternalUserGroupRepo(token, repo.V1)
	externalGroupIDs, err := UserGroupRepo.GetUserBelongingGroupIDs(userID)
	if err != nil {
		return nil, err
	}
	groupIDs = append(groupIDs, externalGroupIDs...)
	return groupIDs, nil
}
