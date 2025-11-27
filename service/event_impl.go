package service

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filters"
)

func (repo *service) CreateEvent(ctx context.Context, params domain.WriteEventParams) (*domain.Event, error) {
	reqID, _ := domain.GetUserID(ctx)
	// groupの確認
	group, err := repo.GetGroup(ctx, params.GroupID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	if !params.TimeStart.Before(params.TimeEnd) {
		return nil, errors.New("event time consistency")
	}

	var r *domain.Room
	if params.RoomID.IsNil() {
		r, err = repo.GormRepo.CreateRoom(domain.CreateRoomArgs{
			WriteRoomParams: domain.WriteRoomParams{
				Place:     params.Place,
				TimeStart: params.TimeStart,
				TimeEnd:   params.TimeEnd,
				Admins:    params.Admins,
			},
			Verified:  false, // イベント専用部屋なので
			CreatedBy: reqID,
		})

		if err != nil {
			return nil, defaultErrorHandling(err)
		}

		params.RoomID = r.ID
	} else {
		r, err = repo.GormRepo.GetRoom(params.RoomID, uuid.Nil)
		if err != nil {
			return nil, defaultErrorHandling(err)
		}
	}

	if !r.ValidateEventTimeAvilability(params.AllowTogether, params.TimeStart, params.TimeEnd) {
		return nil, errors.New("room time consistency")
	}

	p := domain.CreateEventArgs{
		Name:          params.Name,
		Description:   params.Description,
		GroupID:       params.GroupID,
		RoomID:        params.RoomID,
		TimeStart:     params.TimeStart,
		TimeEnd:       params.TimeEnd,
		Admins:        params.Admins,
		Tags:          params.Tags,
		AllowTogether: params.AllowTogether,
		Open:          params.Open,
		CreatedBy:     reqID,
		ID:            uuid.Must(uuid.NewV7()),
	}
	event, err := repo.GormRepo.CreateEvent(p)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	for _, groupMember := range group.Members {
		_ = repo.GormRepo.UpsertEventSchedule(event.ID, groupMember.ID, domain.Pending)
	}
	return repo.GetEvent(ctx, event.ID)
}

func (repo *service) UpdateEvent(ctx context.Context, eventID uuid.UUID, params domain.WriteEventParams) (*domain.Event, error) {
	reqID, _ := domain.GetUserID(ctx)

	if !repo.IsEventAdmins(ctx, eventID) {
		return nil, domain.ErrForbidden
	}

	currentEvent, err := repo.GetEvent(ctx, eventID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	// groupの確認
	group, err := repo.GetGroup(ctx, params.GroupID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	if !params.TimeStart.Before(params.TimeEnd) {
		return nil, errors.New("event time consistency")
	}

	var r *domain.Room
	if params.RoomID.IsNil() {
		r, err = repo.GormRepo.CreateRoom(domain.CreateRoomArgs{
			WriteRoomParams: domain.WriteRoomParams{
				Place:     params.Place,
				TimeStart: params.TimeStart,
				TimeEnd:   params.TimeEnd,
				Admins:    params.Admins,
			},
			Verified:  false, // イベント専用部屋なので
			CreatedBy: reqID,
		})

		if err != nil {
			return nil, defaultErrorHandling(err)
		}

		params.RoomID = r.ID
	} else {
		r, err = repo.GormRepo.GetRoom(params.RoomID, eventID)
		if err != nil {
			return nil, defaultErrorHandling(err)
		}
	}

	if !r.ValidateEventTimeAvilability(params.AllowTogether, params.TimeStart, params.TimeEnd) {
		return nil, errors.New("room time consistency")
	}

	p := domain.UpdateEventArgs{
		Name:          params.Name,
		Description:   params.Description,
		GroupID:       params.GroupID,
		RoomID:        params.RoomID,
		TimeStart:     params.TimeStart,
		TimeEnd:       params.TimeEnd,
		Admins:        params.Admins,
		Tags:          params.Tags,
		AllowTogether: params.AllowTogether,
		Open:          params.Open,
		CreatedBy:     reqID,
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
	return repo.GetEvent(ctx, event.ID)
}

func (repo *service) AddEventTag(ctx context.Context, eventID uuid.UUID, tagName string, locked bool) error {

	if locked && !repo.IsEventAdmins(ctx, eventID) {
		return domain.ErrForbidden
	}
	return repo.GormRepo.AddEventTag(eventID, domain.EventTagParams{
		Name: tagName, Locked: locked,
	})
}

func (repo *service) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	if !repo.IsEventAdmins(ctx, eventID) {
		return domain.ErrForbidden
	}

	return repo.GormRepo.DeleteEvent(eventID)
}

// DeleteTagInEvent delete a tag in that Event
func (repo *service) DeleteEventTag(ctx context.Context, eventID uuid.UUID, tagName string) error {
	deleteLocked := repo.IsEventAdmins(ctx, eventID)

	return repo.GormRepo.DeleteEventTag(eventID, tagName, deleteLocked)
}

func (repo *service) GetEvent(ctx context.Context, eventID uuid.UUID) (*domain.Event, error) {
	event, err := repo.GormRepo.GetEvent(eventID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	// add traQ groups and users
	g, err := repo.GetGroup(ctx, event.Group.ID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	event.Group = *g
	users, err := repo.GetAllUsers(ctx, false, true)
	if err != nil {
		return event, err
	}
	userMap := createUserMap(users)
	c, ok := userMap[event.CreatedBy.ID]
	if ok {
		event.CreatedBy = *c
	}
	for j, eventAdmin := range event.Admins {
		a, ok := userMap[eventAdmin.ID]
		if ok {
			event.Admins[j] = *a
		}
	}

	return event, nil
}

func (repo *service) UpsertMeEventSchedule(ctx context.Context, eventID uuid.UUID, schedule domain.ScheduleStatus) error {
	reqID, _ := domain.GetUserID(ctx)

	event, err := repo.GetEvent(ctx, eventID)
	if err != nil {
		return err
	}
	if !repo.IsGroupMember(ctx, reqID, event.Group.ID) && !event.Open {
		return domain.ErrForbidden
	}

	err = repo.GormRepo.UpsertEventSchedule(eventID, reqID, schedule)
	return defaultErrorHandling(err)
}

func (repo *service) GetEvents(ctx context.Context, expr filters.Expr) ([]*domain.Event, error) {
	reqID, _ := domain.GetUserID(ctx)

	expr = addTraQGroupIDs(repo, reqID, expr)

	es, err := repo.GormRepo.GetAllEvents(expr)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	return es, nil
}

func (repo *service) GetEventsWithGroup(ctx context.Context, expr filters.Expr) ([]*domain.Event, error) {
	reqID, _ := domain.GetUserID(ctx)

	expr = addTraQGroupIDs(repo, reqID, expr)

	events, err := repo.GormRepo.GetAllEvents(expr)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	// add traQ groups and users
	groups, err := repo.GetAllGroups(ctx)
	if err != nil {
		return events, nil
	}
	groupMap := createGroupMap(groups)
	users, err := repo.GetAllUsers(ctx, false, true)
	if err != nil {
		return events, err
	}
	userMap := createUserMap(users)
	for i := range events {
		g, ok := groupMap[events[i].Group.ID]
		if ok {
			events[i].Group = *g
		}
		c, ok := userMap[events[i].CreatedBy.ID]
		if ok {
			events[i].CreatedBy = *c
		}
		for j, eventAdmin := range events[i].Admins {
			a, ok := userMap[eventAdmin.ID]
			if ok {
				events[i].Admins[j] = *a
			}
		}
	}

	return events, nil
}

func (repo *service) IsEventAdmins(ctx context.Context, eventID uuid.UUID) bool {
	reqID, _ := domain.GetUserID(ctx)

	event, err := repo.GormRepo.GetEvent(eventID)
	if err != nil {
		return false
	}
	for _, admin := range event.Admins {
		if reqID == admin.ID {
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
func addTraQGroupIDs(repo *service, userID uuid.UUID, expr filters.Expr) filters.Expr {
	t, err := repo.GormRepo.GetToken(userID)
	if err != nil {
		return expr
	}

	var fixExpr func(filters.Expr) filters.Expr

	fixExpr = func(expr filters.Expr) filters.Expr {
		switch e := expr.(type) {
		case *filters.CmpExpr:
			if e.Attr == filters.AttrBelong {
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
				return &filters.LogicOpExpr{
					LogicOp: filters.Or,
					LHS:     e,
					RHS:     filters.FilterGroupIDs(groupIDs...),
				}
			}
			return e
		case *filters.LogicOpExpr:
			return &filters.LogicOpExpr{
				LogicOp: e.LogicOp,
				LHS:     fixExpr(e.LHS),
				RHS:     fixExpr(e.RHS),
			}
		}
		return nil
	}
	return fixExpr(expr)
}
