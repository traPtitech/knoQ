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
	IsTraQGroup bool      `json:"isTraQGroup"`
	CreatedBy   uuid.UUID `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// EventRes is event response
type EventRes struct {
	repo.Event
	Tags []TagRelationRes `json:"tags"`
}

// TagRelationRes show relation one to tag
type TagRelationRes struct {
	ID     uuid.UUID `json:"id"`
	Locked bool      `json:"locked"`
}

func formatGroupRes(g *repo.Group) (*GroupRes, error) {
	res := new(GroupRes)
	err := copier.Copy(&res, g)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for _, user := range g.Members {
		res.Members = append(res.Members, user.ID)
	}
	return res, nil
}

func formatGroupsRes(g []repo.Group) ([]GroupRes, error) {
	res := []GroupRes{}
	err := copier.Copy(&res, g)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for i, v := range g {
		for _, user := range v.Members {
			res[i].Members = append(res[i].Members, user.ID)
		}
	}
	return res, err
}

func formatEventRes(e *repo.Event) (*EventRes, error) {
	res := new(EventRes)
	err := copier.Copy(&res, e)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return res, err
}

func formatEventsRes(e []repo.Event) ([]EventRes, error) {
	res := []EventRes{}
	err := copier.Copy(&res, e)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return res, err
}
