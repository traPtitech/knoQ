package service

import repo "room/repository"

// Dao DataAccess Object
type Dao struct {
	Repo                      repo.Repository
	InitExternalUserGroupRepo func(token string, ver repo.TraQVersion) interface {
		repo.GroupRepository
		repo.UserBodyRepository
	}
	InitTraPGroupRepo func(token string, ver repo.TraQVersion) interface {
		repo.GroupRepository
	}
	ExternalRoomRepo repo.RoomRepository
}
