package presentation

import "github.com/gofrs/uuid"

//go:generate gotypeconverter -s domain.User -d UserRes -o converter.go .
//go:generate gotypeconverter -s []*domain.User -d []*UserRes -o converter.go .
type UserRes struct {
	ID          uuid.UUID `json:"userId"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Icon        string    `json:"icon"`
	Privileged  bool      `json:"privileged"`
	State       int       `json:"state"`
}
