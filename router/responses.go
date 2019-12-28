package router

import (
	"github.com/gofrs/uuid"
	repo "room/repository"
	"time"
	"github.com/ulule/deepcopier"
)

type GroupRes struct {
	ID uuid.UUID `json:"id"`
	GroupReq
	Members     []string  `json:"members"`
	IsTraQGroup bool      `json:"is_traQ_group"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func formatGroupRes(g *repo.Group) (res *GroupRes, err error) {
	deepcopier.Copy(g).To(res)
	return
}
