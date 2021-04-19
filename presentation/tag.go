package presentation

import "github.com/gofrs/uuid"

type TagReq struct {
	Name string `json:"name"`
}

//go:generate gotypeconverter -s domain.Tag -d TagRes -o converter.go .
//go:generate gotypeconverter -s []*domain.Tag -d []*TagRes -o converter.go .
type TagRes struct {
	ID   uuid.UUID `json:"tagId"`
	Name string    `json:"name"`
	Model
}
