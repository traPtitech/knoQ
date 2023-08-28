package production

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filter"
	"github.com/traPtitech/knoQ/infra/db"
)

func (repo *Repository) CreateEvent(params domain.WriteEventParams, info *domain.ConInfo) (*domain.Event, error) {
	// groupの確認
	group, err := repo.GetGroup(params.GroupID, info)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	p := db.WriteEventParams{
		WriteEventParams: params,
		CreatedBy:        info.ReqUserID,
	}
	event, err := repo.GormRepo.CreateEvent(p)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	for _, groupMember := range group.Members {
		_ = repo.GormRepo.UpsertEventSchedule(event.ID, groupMember.ID, domain.Pending)

	}
	return repo.GetEvent(event.ID, info)
}

func (repo *Repository) UpdateEvent(eventID uuid.UUID, params domain.WriteEventParams, info *domain.ConInfo) (*domain.Event, error) {
	if !repo.IsEventAdmins(eventID, info) {
		return nil, domain.ErrForbidden
	}

	currentEvent, err := repo.GetEvent(eventID, info)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	// groupの確認
	group, err := repo.GetGroup(params.GroupID, info)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	p := db.WriteEventParams{
		WriteEventParams: params,
		CreatedBy:        info.ReqUserID,
	}
	event, err := repo.GormRepo.UpdateEvent(eventID, p)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	for _, groupMember := range group.Members {
		exist := false
		for _, currentAttendee := range currentEvent.Attendees {
			if currentAttendee.UserID == groupMember.ID {
				exist = true
			}
		}
		if !exist {
			_ = repo.GormRepo.UpsertEventSchedule(event.ID, groupMember.ID, domain.Pending)
		}

	}
	return repo.GetEvent(event.ID, info)
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
	if repo.IsEventAdmins(eventID, info) {
		deleteLocked = true
	}
	return repo.GormRepo.DeleteEventTag(eventID, tagName, deleteLocked)
}

func (repo *Repository) GetEvent(eventID uuid.UUID, info *domain.ConInfo) (*domain.Event, error) {
	e, err := repo.GormRepo.GetEvent(eventID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	event := db.ConvEventTodomainEvent(*e)
	// add traQ groups and users
	g, err := repo.GetGroup(e.GroupID, info)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	event.Group = *g
	users, err := repo.GetAllUsers(false, true, info)
	if err != nil {
		return &event, err
	}
	userMap := createUserMap(users)
	c, ok := userMap[e.CreatedByRefer]
	if ok {
		event.CreatedBy = *c
	}
	for j, eventAdmin := range e.Admins {
		a, ok := userMap[eventAdmin.UserID]
		if ok {
			event.Admins[j] = *a
		}
	}

	return &event, nil
}

func (repo *Repository) UpsertMeEventSchedule(eventID uuid.UUID, schedule domain.ScheduleStatus, info *domain.ConInfo) error {
	event, err := repo.GetEvent(eventID, info)
	if err != nil {
		return err
	}
	if !repo.IsGroupMember(info.ReqUserID, event.Group.ID, info) && !event.Open {
		return domain.ErrForbidden
	}
	now := time.Now()
	if event.TimeStart.Before(now) {
		return domain.ErrTimeHasPassed
	}

	err = repo.GormRepo.UpsertEventSchedule(eventID, info.ReqUserID, schedule)
	return defaultErrorHandling(err)
}

func (repo *Repository) GetEvents(expr filter.Expr, info *domain.ConInfo) ([]*domain.Event, error) {
	expr = addTraQGroupIDs(repo, info.ReqUserID, expr)

	es, err := repo.GormRepo.GetAllEvents(expr)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	events := db.ConvSPEventToSPdomainEvent(es)

	// add traQ groups and users
	groups, err := repo.GetAllGroups(info)
	if err != nil {
		return events, nil
	}
	groupMap := createGroupMap(groups)
	users, err := repo.GetAllUsers(false, true, info)
	if err != nil {
		return events, err
	}
	userMap := createUserMap(users)
	for i := range events {
		g, ok := groupMap[es[i].GroupID]
		if ok {
			events[i].Group = *g
		}
		c, ok := userMap[es[i].CreatedByRefer]
		if ok {
			events[i].CreatedBy = *c
		}
		for j, eventAdmin := range es[i].Admins {
			a, ok := userMap[eventAdmin.UserID]
			if ok {
				events[i].Admins[j] = *a
			}
		}
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

func createGroupMap(groups []*domain.Group) map[uuid.UUID]*domain.Group {
	groupMap := make(map[uuid.UUID]*domain.Group)
	for _, group := range groups {
		groupMap[group.ID] = group
	}
	return groupMap
}

func createUserMap(users []*domain.User) map[uuid.UUID]*domain.User {
	userMap := make(map[uuid.UUID]*domain.User)
	for _, user := range users {
		userMap[user.ID] = user
	}
	return userMap
}

// add traQ group and traP(111...)
func addTraQGroupIDs(repo *Repository, userID uuid.UUID, expr filter.Expr) filter.Expr {
	t, err := repo.GormRepo.GetToken(userID)
	if err != nil {
		return expr
	}

	var fixExpr func(filter.Expr) filter.Expr

	fixExpr = func(expr filter.Expr) filter.Expr {
		switch e := expr.(type) {
		case *filter.CmpExpr:
			if e.Attr == filter.AttrBelong {
				id, ok := e.Value.(uuid.UUID)
				if !ok {
					return e
				}
				groupIDs, err := repo.TraQRepo.GetUserBelongingGroupIDs(t, id)
				if err != nil {
					return e
				}
				// add traP
				user, err := repo.GormRepo.GetUser(id)
				if err != nil {
					return e
				}
				if user.Provider.Issuer == traQIssuerName {
					groupIDs = append(groupIDs, traPGroupID)
				}
				return &filter.LogicOpExpr{
					LogicOp: filter.Or,
					Lhs:     e,
					Rhs:     filter.FilterGroupIDs(groupIDs...),
				}
			}
			return e
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
