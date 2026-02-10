package service

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filters"
)

func (s *service) CreateEvent(ctx context.Context, reqID uuid.UUID, params domain.WriteEventParams) (*domain.Event, error) {
	// groupの確認
	group, err := s.GetGroup(ctx, params.GroupID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	p := domain.UpsertEventArgs{
		WriteEventParams: params,
		CreatedBy:        reqID,
	}
	event, err := s.GormRepo.CreateEvent(p)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	for _, groupMember := range group.Members {
		_ = s.GormRepo.UpsertEventSchedule(event.ID, groupMember.ID, domain.Pending)
	}
	return s.GetEvent(ctx, event.ID)
}

func (s *service) UpdateEvent(ctx context.Context, reqID uuid.UUID, eventID uuid.UUID, params domain.WriteEventParams) (*domain.Event, error) {

	if !s.IsEventAdmins(ctx, reqID, eventID) {
		return nil, domain.ErrForbidden
	}

	currentEvent, err := s.GetEvent(ctx, eventID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	// groupの確認
	group, err := s.GetGroup(ctx, params.GroupID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	p := domain.UpsertEventArgs{
		WriteEventParams: params,
		CreatedBy:        reqID,
	}
	// RoomIDの存在を確認 RoomがなくPlaceがあれば新たに作成

	// if params.RoomID == uuid.Nil {
	// 	if params.Place == "" {
	// 		roomParams := domain.WriteRoomParams {
	// 			Place: params.Place,
	// 			TimeStart: params.TimeStart,
	// 			TimeEnd: params.TimeEnd,
	// 			Admins: params.Admins,
	// 		}
	// 		// UnVerifiedを仮定
	// 		_,err = s.CreateUnVerifiedRoom(ctx,reqID,roomParams)
	// 		if err != nil {
	// 			return nil,err
	// 		}
	// 	} else{
	// 		return nil, ErrRoomUndefined
	// 	}
	// }

	// Event Save を使っている
	event, err := s.GormRepo.UpdateEvent(eventID, p)
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
			_ = s.GormRepo.UpsertEventSchedule(event.ID, groupMember.ID, domain.Pending)
		}

	}
	return s.GetEvent(ctx, event.ID)
}

func (s *service) AddEventTag(ctx context.Context, reqID uuid.UUID, eventID uuid.UUID, tagName string, locked bool) error {

	if locked && !s.IsEventAdmins(ctx, reqID, eventID) {
		return domain.ErrForbidden
	}
	return s.GormRepo.AddEventTag(eventID, domain.EventTagParams{
		Name: tagName, Locked: locked,
	})
}

func (s *service) DeleteEvent(ctx context.Context, reqID uuid.UUID, eventID uuid.UUID) error {
	if !s.IsEventAdmins(ctx, reqID, eventID) {
		return domain.ErrForbidden
	}

	return s.GormRepo.DeleteEvent(eventID)
}

// DeleteTagInEvent delete a tag in that Event
func (s *service) DeleteEventTag(ctx context.Context, reqID uuid.UUID, eventID uuid.UUID, tagName string) error {
	deleteLocked := s.IsEventAdmins(ctx, reqID, eventID)

	return s.GormRepo.DeleteEventTag(eventID, tagName, deleteLocked)
}

func (s *service) GetEvent(ctx context.Context, eventID uuid.UUID) (*domain.Event, error) {
	event, err := s.GormRepo.GetEvent(eventID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	// add traQ groups and users
	g, err := s.GetGroup(ctx, event.Group.ID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	event.Group = *g
	users, err := s.GetAllUsers(ctx, false, true)
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

func (s *service) UpsertMeEventSchedule(ctx context.Context, reqID uuid.UUID, eventID uuid.UUID, schedule domain.ScheduleStatus) error {
	event, err := s.GetEvent(ctx, eventID)
	if err != nil {
		return err
	}
	if !s.IsGroupMember(ctx, reqID, event.Group.ID) && !event.Open {
		return domain.ErrForbidden
	}

	err = s.GormRepo.UpsertEventSchedule(eventID, reqID, schedule)
	return defaultErrorHandling(err)
}

func (s *service) GetEvents(ctx context.Context, reqID uuid.UUID, expr filters.Expr) ([]*domain.Event, error) {

	expr = addTraQGroupIDs(s, reqID, expr)

	es, err := s.GormRepo.GetAllEvents(expr)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	return es, nil
}

func (s *service) GetEventsWithGroup(ctx context.Context, reqID uuid.UUID, expr filters.Expr) ([]*domain.Event, error) {
	expr = addTraQGroupIDs(s, reqID, expr)

	events, err := s.GormRepo.GetAllEvents(expr)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	// add traQ groups and users
	groups, err := s.GetAllGroups(ctx)
	if err != nil {
		return events, nil
	}
	groupMap := createGroupMap(groups)
	users, err := s.GetAllUsers(ctx, false, true)
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

func (s *service) IsEventAdmins(ctx context.Context, reqID uuid.UUID, eventID uuid.UUID) bool {
	event, err := s.GormRepo.GetEvent(eventID)
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
func addTraQGroupIDs(s *service, userID uuid.UUID, expr filters.Expr) filters.Expr {
	t, err := s.GormRepo.GetToken(userID)
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
				groupIDs, err := s.TraQRepo.GetUserBelongingGroupIDs(t, id)
				if err != nil {
					return e
				}
				// add traP
				user, err := s.GormRepo.GetUser(id)
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
