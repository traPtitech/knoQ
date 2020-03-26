package router

import (
	repo "room/repository"
	"time"

	"github.com/gofrs/uuid"
)

// GroupReq is group request model
type GroupReq struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	JoinFreely  bool        `json:"open"`
	Members     []uuid.UUID `json:"members"`
}

// RoomReq is group request model
type RoomReq struct {
	Place     string    `json:"place"`
	TimeStart time.Time `json:"timeStart"`
	TimeEnd   time.Time `json:"timeEnd"`
}

type TagReq struct {
	Name string `json:"name"`
}

func formatGroup(req *GroupReq) (g *repo.Group, err error) {
	g = &repo.Group{
		Name:        req.Name,
		Description: req.Description,
		JoinFreely:  req.JoinFreely,
	}

	g.Members = make([]repo.User, 0, len(req.Members))
	for _, v := range req.Members {
		g.Members = append(g.Members, repo.User{ID: v})
	}
	return
}
