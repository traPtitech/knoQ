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

type UserRes struct {
	ID          uuid.UUID `json:"userId"`
	Admin       bool      `json:"admin"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
}

type RoomRes struct {
	ID            uuid.UUID `json:"roomId"`
	Place         string    `json:"place"`
	Public        bool      `json:"public"`
	TimeStart     string    `json:"timeStart"`
	TimeEnd       string    `json:"timeEnd"`
	AvailableTime []repo.StartEndTime
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func formatGroupRes(g *repo.Group, IsTraQgroup bool) *GroupRes {
	res := &GroupRes{
		ID: g.ID,
		GroupReq: GroupReq{
			Name:        g.Name,
			Description: g.Description,
			JoinFreely:  g.JoinFreely,
			Members:     formatGroupMembersRes(g.Members),
		},
		IsTraQGroup: IsTraQgroup,
		CreatedBy:   g.CreatedBy,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
	return res
}

func formatGroupMembersRes(ms []repo.User) []uuid.UUID {
	ids := make([]uuid.UUID, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}

func formatGroupsRes(gs []*repo.Group, IsTraQGroup bool) []*GroupRes {
	res := make([]*GroupRes, len(gs))
	for i, g := range gs {
		res[i] = formatGroupRes(g, IsTraQGroup)
	}
	return res
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

func formatUserRes(u *repo.User) *UserRes {
	return &UserRes{
		ID:          u.ID,
		Admin:       u.Admin,
		Name:        u.Name,
		DisplayName: u.DisplayName,
	}
}

func formatRoomRes(r *repo.Room) *RoomRes {
	return &RoomRes{
		ID:            r.ID,
		Place:         r.Place,
		Public:        r.Public,
		TimeStart:     r.TimeStart.Format(time.RFC3339),
		TimeEnd:       r.TimeEnd.Format(time.RFC3339),
		AvailableTime: r.CalcAvailableTime(),
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}
