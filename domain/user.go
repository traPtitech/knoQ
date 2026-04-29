package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"
)

type User struct {
	ID          uuid.UUID
	Name        string
	DisplayName string
	Icon        string
	Privileged  bool
	State       int

	Provider *Provider
}

type Provider struct {
	Issuer  string
	Subject string // TODO: これは何?
}

type UserService interface {
	GetOAuthURL(ctx context.Context) (url, state, codeVerifier string)
	// LoginUser OAuthによってユーザーを得る
	LoginUser(ctx context.Context, query, state, codeVerifier string) (*User, error)

	GetUser(ctx context.Context, userID uuid.UUID) (*User, error)
	GetUserMe(ctx context.Context, reqID uuid.UUID) (*User, error)
	GetAllUsers(ctx context.Context, includeSuspend, includeBot bool) ([]*User, error)
	// ReplaceToken(userID uuid.UUID, token string) error
	// GetToken(info *ConInfo) (string, error)
	ReNewMyiCalSecret(ctx context.Context, reqID uuid.UUID) (string, error)
	GetMyiCalSecret(ctx context.Context, reqID uuid.UUID) (string, error)

	IsPrivilege(ctx context.Context, reqID uuid.UUID) bool
	GrantPrivilege(ctx context.Context, userID uuid.UUID) error
	SyncUsers(ctx context.Context, reqID uuid.UUID) error
}

type TokenArgs struct {
	AccessToken  string
	TokenType    string
	RefreshToken string
	Expiry       time.Time
}

type ProviderArgs struct {
	Issuer  string
	Subject string
}

type SaveUserArgs struct {
	UserID uuid.UUID
	State  int
	TokenArgs
	ProviderArgs
}

type SyncUserArgs struct {
	UserID uuid.UUID
	State  int
	ProviderArgs
}

type UserRepository interface {
	SaveUser(ctx context.Context, args SaveUserArgs) (*User, error)
	UpdateiCalSecret(ctx context.Context, userID uuid.UUID, secret string) error
	GetUser(ctx context.Context, userID uuid.UUID) (*User, error)
	GetAllUsers(ctx context.Context, onlyActive bool) ([]*User, error)
	SyncUsers(ctx context.Context, args []SyncUserArgs) error
	GrantPrivilege(ctx context.Context, userID uuid.UUID) error
	GetICalSecret(ctx context.Context, userID uuid.UUID) (string, error)
	GetToken(ctx context.Context, userID uuid.UUID) (*oauth2.Token, error)
}
