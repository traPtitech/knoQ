package router

import (
	"fmt"
	repo "room/repository"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/copier"
)

type GroupRes struct {
	ID uuid.UUID `json:"id"`
	GroupReq
	IsTraQGroup bool      `json:"is_traQ_group"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func formatGroupRes(g *repo.Group) (*GroupRes, error) {
	res := new(GroupRes)
	err := copier.Copy(&res, g)
	if err != nil {
		fmt.Println(err)
	}
	for _, user := range g.Members {
		res.Members = append(res.Members, user.ID)
	}
	return res, err
}

func formatGroupsRes(g []repo.Group) ([]GroupRes, error) {
	res := []GroupRes{}
	err := copier.Copy(&res, g)
	if err != nil {
		fmt.Println(err)
	}
	for i, v := range g {
		for _, user := range v.Members {
			res[i].Members = append(res[i].Members, user.ID)
		}

	}
	return res, err

}
