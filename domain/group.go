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
}
