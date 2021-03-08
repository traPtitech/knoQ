package domain

import "github.com/gofrs/uuid"

type WriteGroupParams struct {
	Name        string
	Description string
	JoinFreely  bool
	Members     []uuid.UUID
	Admins      []uuid.UUID
}

type Group struct {
	ID          uuid.UUID
	Name        string
	Description string
	JoinFreely  bool
	Members     []User
	Admins      []User
	CreatedBy   User
}
