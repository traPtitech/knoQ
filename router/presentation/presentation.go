package presentation

import "time"

type Model struct {
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

type Version struct {
	Version  string `json:"version"`
	Revision string `json:"revision"`
}
