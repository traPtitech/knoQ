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
	SaveUser(args SaveUserArgs) (*User, error)
	UpdateiCalSecret(userID uuid.UUID, secret string) error
	GetUser(userID uuid.UUID) (*User, error)
	GetAllUsers(onlyActive bool) ([]*User, error)
	SyncUsers(args []SyncUserArgs) error
	GrantPrivilege(userID uuid.UUID) error
	GetICalSecret(userID uuid.UUID) (string, error)
	GetToken(userID uuid.UUID) (*oauth2.Token, error)
}
