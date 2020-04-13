package service

import (
	repo "room/repository"

	"github.com/gofrs/uuid"
)

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

func (d Dao) CreateGroup(token string, groupParams repo.WriteGroupParams) (*GroupRes, error) {
	UserGroupRepo := d.InitExternalUserGroupRepo(token, repo.V3)
	allUsers, err := UserGroupRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}

	if err := groupMembersValidation(groupParams, allUsers); err != nil {
		return nil, err
	}

	group, err := d.Repo.CreateGroup(groupParams)
	return FormatGroupRes(group, false), err
}

func (d Dao) UpdateGroup(token string, groupID uuid.UUID, groupParams repo.WriteGroupParams) (*GroupRes, error) {
	UserGroupRepo := d.InitExternalUserGroupRepo(token, repo.V3)
	allUsers, err := UserGroupRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}

	if err := groupMembersValidation(groupParams, allUsers); err != nil {
		return nil, err
	}

	group, err := d.Repo.UpdateGroup(groupID, groupParams)
	return FormatGroupRes(group, false), err

}

func (d Dao) AddUserToGroup(token string, groupID uuid.UUID, userID uuid.UUID) error {
	UserGroupRepo := d.InitExternalUserGroupRepo(token, repo.V3)
	allUsers, err := UserGroupRepo.GetAllUsers()
	if err != nil {
		return err
	}

	exist := false
	for _, user := range allUsers {
		if userID == user.ID {
			exist = true
		}
	}
	if !exist {
		return repo.ErrInvalidArg
	}
	return d.Repo.AddUserToGroup(groupID, userID)
}

func groupMembersValidation(groupParams repo.WriteGroupParams, allUsers []*repo.User) error {
	// member validation
	for _, paramUserID := range groupParams.Members {
		exist := false
		for _, user := range allUsers {
			if paramUserID == user.ID {
				exist = true
			}
		}
		if !exist {
			return repo.ErrInvalidArg
		}
	}
	return nil
}
