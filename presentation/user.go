package presentation

import "github.com/gofrs/uuid"

type UserRes struct {
	ID          uuid.UUID `json:"userId"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Icon        string    `json:"icon"`
	Privileged  bool      `json:"privileged"`
	State       int       `json:"state"`
}
