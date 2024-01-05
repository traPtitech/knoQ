package presentation

import (
	"github.com/gofrs/uuid"
)

//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s GroupReq -d domain.WriteGroupParams -o converter.go .
type GroupReq struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	JoinFreely  bool        `json:"open"`
	Members     []uuid.UUID `json:"members"`
	Admins      []uuid.UUID `json:"admins"`
}

//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s domain.Group -d GroupRes -o converter.go .
//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s []*domain.Group -d []*GroupRes -o converter.go .
type GroupRes struct {
	ID uuid.UUID `json:"groupId"`
	GroupReq
	IsTraQGroup bool      `json:"isTraQGroup"`
	CreatedBy   uuid.UUID `json:"createdBy"`
	Model
}
