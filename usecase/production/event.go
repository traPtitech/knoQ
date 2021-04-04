package production

import (
	"errors"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filter"
	"github.com/traPtitech/knoQ/infra/db"
	traQ "github.com/traPtitech/traQ/router/v3"
)

func (repo *Repository) CreateEvent(params domain.WriteEventParams, info *domain.ConInfo) (*domain.Event, error) {
	// groupの確認
	_, err := repo.GetGroup(params.GroupID, info)
	if err != nil {
		return nil, err
	}

	p := db.WriteEventParams{
		WriteEventParams: params,
		CreatedBy:        info.ReqUserID,
	}
	event, err := repo.gormRepo.CreateEvent(p)
	if err != nil {
		return nil, err
	}
	e := db.ConvertEventTodomainEvent(*event)
	return &e, nil
}

func (repo *Repository) UpdateEvent(eventID uuid.UUID, params domain.WriteEventParams, info *domain.ConInfo) (*domain.Event, error) {
	// groupの確認
	_, err := repo.GetGroup(params.GroupID, info)
	if err != nil {
		return nil, err
	}

	p := db.WriteEventParams{
		WriteEventParams: params,
		CreatedBy:        info.ReqUserID,
	}
	event, err := repo.gormRepo.UpdateEvent(eventID, p)
	if err != nil {
		return nil, err
	}
	e := db.ConvertEventTodomainEvent(*event)
	return &e, nil
}

func (repo *Repository) AddTagToEvent(eventID uuid.UUID, tagName string, locked bool, info *domain.ConInfo) error {
	if locked && !repo.IsEventAdmins(eventID, info) {
		return errors.New("Forbidden")
	}
	return repo.gormRepo.AddEventTag(eventID, domain.EventTagParams{
		Name: tagName, Locked: locked,
	})
}

func (repo *Repository) DeleteEvent(eventID uuid.UUID, info *domain.ConInfo) error {
	if !repo.IsEventAdmins(eventID, info) {
		return errors.New("Forbidden")
	}

	return repo.gormRepo.DeleteEvent(eventID)
}

// DeleteTagInEvent delete a tag in that Event
func (repo *Repository) DeleteTagInEvent(eventID uuid.UUID, tagName string, info *domain.ConInfo) error {
	deleteLocked := false
	if !repo.IsEventAdmins(eventID, info) {
		deleteLocked = true
	}
	return repo.gormRepo.DeleteEventTag(eventID, tagName, deleteLocked)
}

func (repo *Repository) GetEvent(eventID uuid.UUID, info *domain.ConInfo) (*domain.Event, error) {
	e, err := repo.gormRepo.GetEvent(eventID)
	if err != nil {
		return nil, err
	}
	event := db.ConvertEventTodomainEvent(*e)
	g, err := repo.GetGroup(event.Group.ID, info)
	if err != nil {
		return nil, err
	}
	event.Group = *g
	return &event, nil
}

func (repo *Repository) GetEvents(expr filter.Expr, info *domain.ConInfo) ([]*domain.Event, error) {
	expr = addTraQGroupIDs(repo, info.ReqUserID, expr)

	es, err := repo.gormRepo.GetAllEvents(expr)
	if err != nil {
		return nil, err
	}
	events := db.ConvertSlicePointerEventToSlicePointerdomainEvent(es)
	t, err := repo.gormRepo.GetToken(info.ReqUserID)
	if err != nil {
		return nil, err
	}
	traQgroups, err := repo.traQRepo.GetAllGroups(t)
	if err != nil {
		return events, nil
	}
	groupMap := traQGroupMap(traQgroups)

	for i, e := range events {
		g, ok := groupMap[e.Group.ID]
		if !ok {
			continue
		}
		events[i].Group = Convertv3UserGroupTodomainGroup(*g)

	}
	return events, nil
}

func (repo *Repository) IsEventAdmins(eventID uuid.UUID, info *domain.ConInfo) bool {
	event, err := repo.gormRepo.GetEvent(eventID)
	if err != nil {
		return false
	}
	for _, admin := range event.Admins {
		if info.ReqUserID == admin.UserID {
			return true
		}
	}
	return false
}

//go:generate gotypeconverter -s v3.UserGroup -d domain.Group -o converter.go .

func traQGroupMap(groups []*traQ.UserGroup) map[uuid.UUID]*traQ.UserGroup {
	groupMap := make(map[uuid.UUID]*traQ.UserGroup)
	for _, group := range groups {
		groupMap[group.ID] = group
	}
	return groupMap
}

func addTraQGroupIDs(repo *Repository, userID uuid.UUID, expr filter.Expr) filter.Expr {
	t, err := repo.gormRepo.GetToken(userID)
	if err != nil {
		return expr
	}

	var fixExpr func(filter.Expr) filter.Expr

	fixExpr = func(expr filter.Expr) filter.Expr {
		switch e := expr.(type) {
		case *filter.CmpExpr:
			if e.Attr == filter.User {
				id, ok := e.Value.(uuid.UUID)
				if !ok {
					return e
				}
				groupIDs, err := repo.traQRepo.GetUserBelongingGroupIDs(t, id)
				if err != nil {
					return e
				}
				return &filter.LogicOpExpr{
					LogicOp: filter.Or,
					Lhs:     e,
					Rhs:     filter.FilterGroupIDs(groupIDs),
				}
			}
		case *filter.LogicOpExpr:
			return &filter.LogicOpExpr{
				LogicOp: e.LogicOp,
				Lhs:     fixExpr(e.Lhs),
				Rhs:     fixExpr(e.Rhs),
			}
		}
		return nil
	}
	return fixExpr(expr)
}
