package infra

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/go-traq"
	"github.com/traPtitech/knoQ/domain"
	"golang.org/x/oauth2"
)

type TraqRepository interface {
	GetOAuthURL() (url, state, codeVerifier string)
	GetOAuthToken(query, state, codeVerifier string) (*oauth2.Token, error)

	GetUser(userID uuid.UUID) (*TraqUserResponse, error)
	GetUsers(includeSuspended bool) ([]*TraqUserResponse, error)
	GetUserMe(accessToken string) (*TraqUserResponse, error)

	GetGroup(groupID uuid.UUID) (*TraqUserGroupResponse, error)
	GetAllGroups() ([]*TraqUserGroupResponse, error)
	GetUserBelongingGroupIDs(accessToken string, userID uuid.UUID) ([]uuid.UUID, error)
	GetGradeGroups() ([]*TraqUserGroupResponse, error)
}

type SaveUserArgs struct{}

type SyncUserArgs struct{}

type UserRepository interface {
	SaveUser(args SaveUserArgs) (*domain.User, error)
	UpdateiCalSecret(userID uuid.UUID, secret string) error
	GetUser(userID uuid.UUID) (*domain.User, error)
	GetAllUsers(onlyActive bool) ([]*domain.User, error)
	SyncUsers(args SyncUserArgs) error
}

type GroupRepository interface{}

type TagRepository interface{}

type RoomRepository interface{}

type EventRepository interface{}

type (
	TraqUserResponse struct {
		ID          uuid.UUID
		Name        string
		DisplayName string
		IconURL     string
		Bot         bool
		State       traq.UserAccountState
		UpdatedAt   time.Time
	}

	TraqUserGroupMember struct {
		ID   uuid.UUID
		Role string
	}

	TraqUserGroupResponse struct {
		ID          uuid.UUID
		Name        string
		Description string
		Type        string
		IconID      uuid.UUID
		Members     []TraqUserGroupMember
		CreatedAt   time.Time
		UpdatedAt   time.Time
		Admins      []uuid.UUID
	}
)
