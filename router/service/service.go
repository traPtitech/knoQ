package service

import repo "github.com/traPtitech/knoQ/repository"

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
