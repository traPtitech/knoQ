package domain

import "github.com/gofrs/uuid"

type User struct {
	ID          uuid.UUID
	Name        string
	DisplayName string
	Icon        string
	Privileged  bool
	State       int
}

type UserRepository interface {
	GetOAuthURL() (url, state, codeVerifier string)
	// LoginUser OAuthによってユーザーを得る
	LoginUser(query, state, codeVerifier string) (*User, error)

	GetUser(userID uuid.UUID, info *ConInfo) (*User, error)
	GetUserMe(info *ConInfo) (*User, error)
	GetAllUsers(includeSuspend, includeBot bool, info *ConInfo) ([]*User, error)
	//ReplaceToken(userID uuid.UUID, token string, info *ConInfo) error
	//GetToken(info *ConInfo) (string, error)
	ReNewMyiCalSecret(info *ConInfo) (string, error)
	GetMyiCalSecret(info *ConInfo) (string, error)

	IsPrevilege(info *ConInfo) bool
	GrantPrivilege(userID uuid.UUID) error
}
