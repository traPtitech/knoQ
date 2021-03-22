package presentation

import "github.com/gofrs/uuid"

//go:generate gotypeconverter -s domain.Tag -d TagRes -o converter.go .
type TagRes struct {
	ID   uuid.UUID `json:"tagId"`
	Name string    `json:"name"`
	Model
}
