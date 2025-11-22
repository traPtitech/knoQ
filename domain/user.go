package domain

import (
	"context"

	"github.com/gofrs/uuid"
)

type User struct {
	ID          uuid.UUID
	Name        string
	DisplayName string
	Icon        string
	Privileged  bool
	State       int
}

type UserService interface {
	GetOAuthURL(ctx context.Context) (url, state, codeVerifier string)
	// LoginUser OAuthによってユーザーを得る
	LoginUser(ctx context.Context, query, state, codeVerifier string) (*User, error)

	GetUser(ctx context.Context, userID uuid.UUID) (*User, error)
	GetUserMe(ctx context.Context) (*User, error)
	GetAllUsers(ctx context.Context, includeSuspend, includeBot bool) ([]*User, error)
	// ReplaceToken(userID uuid.UUID, token string) error
	// GetToken(info *ConInfo) (string, error)
	ReNewMyiCalSecret(ctx context.Context) (string, error)
	GetMyiCalSecret(ctx context.Context) (string, error)

	IsPrivilege(ctx context.Context) bool
	GrantPrivilege(ctx context.Context, userID uuid.UUID) error
	SyncUsers(ctx context.Context) error
}
