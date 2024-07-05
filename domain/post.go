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

type PostRepository interface {
	CreatePost(params WritePostParams) (*Post, error)
	GetPost(MessageID uuid.UUID) (*Post, error)
}
