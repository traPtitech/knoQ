package domain

import "github.com/gofrs/uuid"

type User struct {
	ID          uuid.UUID
	Name        string
	DisplayName string
	Privileged  bool
	IsTrap      bool
}

type WriteUserParams struct {
	UserIdentifier string
	Name           string
	DisplayName    string // option
	Icon           string // option
	//Provider       string // never
}

type UserRepository interface {
	// LoginUser OAuthによってユーザーを得る
	LoginUser(code, codeVerifier string) (*User, error)
	GetUser(userID uuid.UUID, info *ConInfo) (*User, error)
	GetAllUsers(*ConInfo) ([]*User, error)
	//ReplaceToken(userID uuid.UUID, token string, info *ConInfo) error
	//GetToken(info *ConInfo) (string, error)
	ReplaceMyiCalSecret(secret string, info *ConInfo) error
	GetMyiCalSecret(info *ConInfo) (string, error)
}
