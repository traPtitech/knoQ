package domain

import "github.com/gofrs/uuid"

type User struct {
	ID         uuid.UUID
	Privileged bool
	IsTrap     bool
}
