package presentation

import "github.com/gofrs/uuid"

type UserRes struct {
	ID          uuid.UUID `json:"userId"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Privileged  bool      `json:"privileged"`
	IsTrap      bool      `json:"isTrap"`
}
