package router

import (
	repo "room/repository"

	"github.com/gofrs/uuid"
)

// GroupReq is group request model
type GroupReq struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	JoinFreely  bool        `json:"open"`
	Members     []uuid.UUID `json:"members"`
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
