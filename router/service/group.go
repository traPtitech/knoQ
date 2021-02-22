package service

import (
	"fmt"
	repo "room/repository"

	"github.com/gofrs/uuid"
)

func (d Dao) GetGroup(token string, groupID uuid.UUID) (*GroupRes, error) {
	ch := make(chan *GroupRes, 3) // 並列数と対応させること
	UserGroupRepo := d.InitExternalUserGroupRepo(token, repo.TraQv3)
	TraPGroupRepo := d.InitTraPGroupRepo(token, repo.TraQv3)
	go func(out chan *GroupRes) {
		group, err := d.Repo.GetGroup(groupID)
		if err != nil {
			out <- nil
			return
		}
		out <- FormatGroupRes(group, false)
	}(ch)
	go func(out chan *GroupRes) {
		group, err := UserGroupRepo.GetGroup(groupID)
		if err != nil {
			out <- nil
			return
		}
		out <- FormatGroupRes(group, true)
	}(ch)
	go func(out chan *GroupRes) {
		group, err := TraPGroupRepo.GetGroup(groupID)
		if err != nil {
			out <- nil
			return
		}
		out <- FormatGroupRes(group, true)
	}(ch)
	for i := 0; i < cap(ch); i++ {
		group := <-ch
		if group != nil {
			return group, nil
		}
	}
	return nil, repo.ErrNotFound
}

func (d Dao) GetUserBelongingGroupIDs(token string, userID uuid.UUID) ([]uuid.UUID, error) {
	ch := make(chan []uuid.UUID, 3) // 並列数と対応させること
	UserGroupRepo := d.InitExternalUserGroupRepo(token, repo.TraQv1)
	TraPGroupRepo := d.InitTraPGroupRepo(token, repo.TraQv3)

	go func(out chan []uuid.UUID) {
		groupIDs, _ := d.Repo.GetUserBelongingGroupIDs(userID)
		out <- groupIDs
	}(ch)
	go func(out chan []uuid.UUID) {
		groupIDs, _ := UserGroupRepo.GetUserBelongingGroupIDs(userID)
		fmt.Println(groupIDs)
		out <- groupIDs
	}(ch)
	go func(out chan []uuid.UUID) {
		groupIDs, _ := TraPGroupRepo.GetUserBelongingGroupIDs(userID)
		out <- groupIDs
	}(ch)

	groupIDs := make([]uuid.UUID, 0)
	for i := 0; i < cap(ch); i++ {
		IDs := <-ch
		groupIDs = append(groupIDs, IDs...)
	}

	return groupIDs, nil
}

func (d Dao) GetAllGroups(token string) ([]*GroupRes, error) {
	ch := make(chan []*GroupRes, 3) // 並列数と対応させること
	UserGroupRepo := d.InitExternalUserGroupRepo(token, repo.TraQv3)
	TraPGroupRepo := d.InitTraPGroupRepo(token, repo.TraQv3)

	go func(out chan []*GroupRes) {
		groups, _ := d.Repo.GetAllGroups()
		out <- FormatGroupsRes(groups, false)
	}(ch)
	go func(out chan []*GroupRes) {
		groups, _ := UserGroupRepo.GetAllGroups()
		out <- FormatGroupsRes(groups, true)
	}(ch)
	go func(out chan []*GroupRes) {
		groups, _ := TraPGroupRepo.GetAllGroups()
		out <- FormatGroupsRes(groups, true)
	}(ch)

	groups := make([]*GroupRes, 0)
	for i := 0; i < cap(ch); i++ {
		gs := <-ch
		groups = append(groups, gs...)
	}
	return groups, nil
}

func (d Dao) CreateGroup(token string, groupParams repo.WriteGroupParams) (*GroupRes, error) {
	UserGroupRepo := d.Repo
	allUsers, err := UserGroupRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}

	if err := groupMembersValidation(groupParams, allUsers); err != nil {
		return nil, err
	}
	if len(groupParams.Admins) == 0 {
		return nil, repo.ErrInvalidArg
	}

	group, err := d.Repo.CreateGroup(groupParams)
	return FormatGroupRes(group, false), err
}

func (d Dao) UpdateGroup(token string, groupID uuid.UUID, groupParams repo.WriteGroupParams) (*GroupRes, error) {
	UserGroupRepo := d.Repo
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
	UserGroupRepo := d.InitExternalUserGroupRepo(token, repo.TraQv3)
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

func groupMembersValidation(groupParams repo.WriteGroupParams, allUsers []*repo.UserMeta) error {
	// member validation
	existUserID := func(ids []uuid.UUID) error {
		for _, paramUserID := range ids {
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
	if existUserID(groupParams.Members) != nil || existUserID(groupParams.Admins) != nil {
		return repo.ErrInvalidArg
	}
	return nil
}
