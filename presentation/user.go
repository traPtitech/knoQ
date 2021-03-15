package presentation

import "github.com/gofrs/uuid"

//go:generate go run github.com/fuji8/gotypeconverter/cmd/type-converter -s domain.User -d domain.WriteUserParams -o converter.go .
type UserRes struct {
	ID          uuid.UUID `json:"userId"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Privileged  bool      `json:"privileged"`
	IsTrap      bool      `json:"isTrap"`
}
