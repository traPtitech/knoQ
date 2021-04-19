package presentation

import (
	"github.com/gofrs/uuid"
)

//go:generate gotypeconverter -s GroupReq -d domain.WriteGroupParams -o converter.go .
type GroupReq struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	JoinFreely  bool        `json:"open"`
	Members     []uuid.UUID `json:"members"`
	Admins      []uuid.UUID `json:"admins"`
}

//go:generate gotypeconverter -s domain.Group -d GroupRes -o converter.go .
//go:generate gotypeconverter -s []*domain.Group -d []*GroupRes -o converter.go .
type GroupRes struct {
	ID uuid.UUID `json:"groupId"`
	GroupReq
	IsTraQGroup bool      `json:"isTraQGroup"`
	CreatedBy   uuid.UUID `json:"createdBy"`
	Model
}
