package domain

import (
	"context"

	"github.com/gofrs/uuid"
)

type Group struct {
	ID          uuid.UUID
	Name        string
	Description string
	JoinFreely  bool
	Members     []User
	Admins      []User
	IsTraQGroup bool `cvt:"->"`
	CreatedBy   User
	Model
}

func (g *Group) AdminsValidation() bool {
	return len(g.Admins) != 0
}

type WriteGroupParams struct {
	Name        string
	Description string
	JoinFreely  bool
	Members     []uuid.UUID
	Admins      []uuid.UUID
}

type GroupService interface {
	CreateGroup(ctx context.Context, groupParams WriteGroupParams) (*Group, error)
	UpdateGroup(ctx context.Context, groupID uuid.UUID, groupParams WriteGroupParams) (*Group, error)
	// AddMeToGroup add me to that group if that group is open.
	AddMeToGroup(ctx context.Context, groupID uuid.UUID) error
	DeleteGroup(ctx context.Context, groupID uuid.UUID) error
	// DeleteMeGroup delete me in that group if that group is open.
	DeleteMeGroup(ctx context.Context, groupID uuid.UUID) error

	GetGroup(ctx context.Context, groupID uuid.UUID) (*Group, error)
	GetAllGroups(ctx context.Context) ([]*Group, error)
	GetUserBelongingGroupIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	GetUserAdminGroupIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	IsGroupAdmins(ctx context.Context, groupID uuid.UUID) bool
	GetGradeGroupNames(ctx context.Context) ([]string, error)
}
