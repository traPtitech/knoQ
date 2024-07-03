package domain

import "github.com/gofrs/uuid"

type Post struct {
	MessageID uuid.UUID
	EventID   uuid.UUID
}

type WritePostParams struct {
	MessageID uuid.UUID
	EventID   uuid.UUID
}
