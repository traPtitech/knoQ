package domain

import "github.com/gofrs/uuid"

type Tag struct {
	ID   uuid.UUID
	Name string
}
