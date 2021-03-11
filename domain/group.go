package domain

import "github.com/gofrs/uuid"

type Group struct {
	ID          uuid.UUID
	Name        string
	Description string
	JoinFreely  bool
	Members     []User
	Admins      []User
	CreatedBy   User
	Model
}

type WriteGroupParams struct {
	Name        string
	Description string
	JoinFreely  bool
	Members     []uuid.UUID
	Admins      []uuid.UUID
}

type GroupRepository interface {
	CreateGroup(groupParams WriteGroupParams, info *ConInfo) (*Group, error)
	UpdateGroup(groupID uuid.UUID, groupParams WriteGroupParams, info *ConInfo) (*Group, error)
	// AddMeToGroup add me to that group if that group is open.
	AddMeToGroup(groupID uuid.UUID, info *ConInfo) error
	DeleteGroup(groupID uuid.UUID, info *ConInfo) error
	// DeleteMeGroup delete me in that group if that group is open.
	DeleteMeGroup(groupID uuid.UUID, info *ConInfo) error

	GetGroup(groupID uuid.UUID, info *ConInfo) (*Group, error)
	GetAllGroups(info *ConInfo) ([]*Group, error)
	GetUserBelongingGroupIDs(userID uuid.UUID, info *ConInfo) ([]uuid.UUID, error)
}
