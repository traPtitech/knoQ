package production

import (
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
	event, err := repo.GormRepo.CreateEvent(p)
	if err != nil {
		return nil, err
	}
	e := db.ConvEventTodomainEvent(*event)
	return &e, nil
}

func (repo *Repository) UpdateEvent(eventID uuid.UUID, params domain.WriteEventParams, info *domain.ConInfo) (*domain.Event, error) {
	if !repo.IsEventAdmins(eventID, info) {
		return nil, domain.ErrForbidden
	}
	// groupの確認
	_, err := repo.GetGroup(params.GroupID, info)
	if err != nil {
		return nil, err
	}

	p := db.WriteEventParams{
		WriteEventParams: params,
		CreatedBy:        info.ReqUserID,
	}
	event, err := repo.GormRepo.UpdateEvent(eventID, p)
	if err != nil {
		return nil, err
	}
	e := db.ConvEventTodomainEvent(*event)
	return &e, nil
}

func (repo *Repository) AddEventTag(eventID uuid.UUID, tagName string, locked bool, info *domain.ConInfo) error {
	if locked && !repo.IsEventAdmins(eventID, info) {
		return domain.ErrForbidden
	}
	return repo.GormRepo.AddEventTag(eventID, domain.EventTagParams{
		Name: tagName, Locked: locked,
	})
}

func (repo *Repository) DeleteEvent(eventID uuid.UUID, info *domain.ConInfo) error {
	if !repo.IsEventAdmins(eventID, info) {
		return domain.ErrForbidden
	}

	return repo.GormRepo.DeleteEvent(eventID)
}

// DeleteTagInEvent delete a tag in that Event
func (repo *Repository) DeleteEventTag(eventID uuid.UUID, tagName string, info *domain.ConInfo) error {
	deleteLocked := false
	if !repo.IsEventAdmins(eventID, info) {
		deleteLocked = true
	}
	return repo.GormRepo.DeleteEventTag(eventID, tagName, deleteLocked)
}

func (repo *Repository) GetEvent(eventID uuid.UUID, info *domain.ConInfo) (*domain.Event, error) {
	e, err := repo.GormRepo.GetEvent(eventID)
	if err != nil {
		return nil, err
	}
	event := db.ConvEventTodomainEvent(*e)
	g, err := repo.GetGroup(event.Group.ID, info)
	if err != nil {
		return nil, err
	}
	event.Group = *g
	return &event, nil
}

func (repo *Repository) GetEvents(expr filter.Expr, info *domain.ConInfo) ([]*domain.Event, error) {
	expr = addTraQGroupIDs(repo, info.ReqUserID, expr)

	es, err := repo.GormRepo.GetAllEvents(expr)
	if err != nil {
		return nil, err
	}
	events := db.ConvSPEventToSPdomainEvent(es)
	t, err := repo.GormRepo.GetToken(info.ReqUserID)
	if err != nil {
		return nil, err
	}
	traQgroups, err := repo.TraQRepo.GetAllGroups(t)
	if err != nil {
		return events, nil
	}
	groupMap := traQGroupMap(traQgroups)

	for i, e := range events {
		g, ok := groupMap[e.Group.ID]
		if !ok {
			continue
		}
		events[i].Group = Convv3UserGroupTodomainGroup(*g)

	}
	return events, nil
}

func (repo *Repository) IsEventAdmins(eventID uuid.UUID, info *domain.ConInfo) bool {
	event, err := repo.GormRepo.GetEvent(eventID)
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
	t, err := repo.GormRepo.GetToken(userID)
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
				groupIDs, err := repo.TraQRepo.GetUserBelongingGroupIDs(t, id)
				if err != nil {
					return e
				}
				return &filter.LogicOpExpr{
					LogicOp: filter.Or,
					Lhs:     e,
					Rhs:     filter.FilterGroupIDs(groupIDs...),
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
