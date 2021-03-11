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
	Name        string
	DisplayName string
	Privileged  bool
	IsTrap      bool
}

type UserRepository interface {
	CreateUser(WriteUserParams, *ConInfo) (*User, error)
	GetUser(userID uuid.UUID, info *ConInfo) (*User, error)
	GetAllUsers(*ConInfo) ([]*User, error)
	ReplaceToken(userID uuid.UUID, token string, info *ConInfo) error
	GetToken(info *ConInfo) (string, error)
	ReplaceMyiCalSecret(secret string, info *ConInfo) error
	GetMyiCalSecret(info *ConInfo) (string, error)
}
