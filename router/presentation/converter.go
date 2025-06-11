package presentation

import (
	"database/sql"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/traPtitech/knoQ/domain"
)

func ConvEventReqWriteTodomainWriteEventParams(src EventReqWrite) (dst domain.WriteEventParams) {
	dst.Name = src.Name
	dst.Description = src.Description
	dst.GroupID = src.GroupID
	if src.RoomID.IsNil() { // not room event
		dst.Venue = sql.NullString{Valid: true, String: src.Place}
	} else {
		dst.IsRoomEvent = true
		dst.RoomID = uuid.NullUUID{Valid: true, UUID: src.RoomID}
	}

	dst.TimeStart = src.TimeStart
	dst.TimeEnd = src.TimeEnd
	dst.Admins = src.Admins
	dst.Tags = make([]domain.EventTagParams, len(src.Tags))
	for i := range src.Tags {
		dst.Tags[i] = domain.EventTagParams(src.Tags[i])
	}
	dst.AllowTogether = src.AllowTogether
	dst.Open = src.Open
	return
}

func ConvGroupReqTodomainWriteGroupParams(src GroupReq) (dst domain.WriteGroupParams) {
	dst = domain.WriteGroupParams(src)
	return
}

func ConvRoomReqTodomainWriteRoomParams(src RoomReq) (dst domain.WriteRoomParams) {
	dst = domain.WriteRoomParams(src)
	return
}

func ConvDomainEventsToEventsResElems(src []*domain.Event) []EventsResElement {
	return lo.Map(src, func(e *domain.Event, _ int) EventsResElement {
		roomID := uuid.Nil
		place := e.Venue.String
		if e.IsRoomEvent {
			roomID = e.Room.ID
			place = e.Room.Name
		}

		return EventsResElement{
			ID:            e.ID,
			Name:          e.Name,
			Description:   e.Description,
			AllowTogether: e.AllowTogether,
			TimeStart:     e.TimeStart,
			TimeEnd:       e.TimeEnd,
			RoomID:        roomID,
			GroupID:       e.Group.ID,
			Place:         place,
			Admins: lo.Map(e.Admins, func(a domain.User, _ int) uuid.UUID {
				return a.ID
			}),
			Tags: lo.Map(e.Tags, func(t domain.EventTag, _ int) EventTagRes {
				return EventTagRes{ID: t.Tag.ID, Name: t.Tag.Name, Locked: t.Locked}
			}),
			CreatedBy: e.CreatedBy.ID,
			Open:      e.Open,
			Attendees: lo.Map(e.Attendees, func(a domain.Attendee, _ int) uuid.UUID {
				return a.UserID
			}),
			Model: Model(e.Model),
		}
	})
}

func ConvSPdomainGroupToSPGroupRes(src []*domain.Group) (dst []*GroupRes) {
	dst = make([]*GroupRes, len(src))
	for i := range src {
		if src[i] != nil {
			dst[i] = new(GroupRes)
			(*dst[i]) = convdomainGroupToGroupRes((*src[i]))
		}
	}
	return
}

func ConvSPdomainTagToSPTagRes(src []*domain.Tag) (dst []*TagRes) {
	dst = make([]*TagRes, len(src))
	for i := range src {
		if src[i] != nil {
			dst[i] = new(TagRes)
			(*dst[i]) = convdomainTagToTagRes((*src[i]))
		}
	}
	return
}

func ConvSPdomainUserToSPUserRes(src []*domain.User) (dst []*UserRes) {
	dst = make([]*UserRes, len(src))
	for i := range src {
		if src[i] != nil {
			dst[i] = new(UserRes)
			(*dst[i]) = convdomainUserToUserRes((*src[i]))
		}
	}
	return
}

func ConvSdomainStartEndTimeToSStartEndTime(src []domain.StartEndTime) (dst []StartEndTime) {
	dst = make([]StartEndTime, len(src))
	for i := range src {
		dst[i] = convdomainStartEndTimeToStartEndTime(src[i])
	}
	return
}

func ConvdomainGroupToGroupRes(src domain.Group) (dst GroupRes) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.Description = src.Description
	dst.JoinFreely = src.JoinFreely
	dst.Members = make([]uuid.UUID, len(src.Members))
	for i := range src.Members {
		dst.Members[i] = convdomainUserTouuidUUID(src.Members[i])
	}
	dst.Admins = make([]uuid.UUID, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = convdomainUserTouuidUUID(src.Admins[i])
	}
	dst.IsTraQGroup = src.IsTraQGroup
	dst.CreatedBy = convdomainUserTouuidUUID(src.CreatedBy)
	dst.Model = Model(src.Model)
	return
}

func ConvdomainTagToTagRes(src domain.Tag) (dst TagRes) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.Model = Model(src.Model)
	return
}

func ConvdomainUserToUserRes(src domain.User) (dst UserRes) {
	dst = UserRes(src)
	return
}

func convdomainAttendeeToEventAttendeeRes(src domain.Attendee) (dst EventAttendeeRes) {
	dst.ID = src.UserID
	dst.Schedule = convdomainScheduleStatusToScheduleStatus(src.Schedule)
	return
}

func convdomainEventTagToEventTagRes(src domain.EventTag) (dst EventTagRes) {
	dst.ID = convdomainTagTouuidUUID(src.Tag)
	dst.Name = src.Tag.Name
	dst.Locked = src.Locked
	return
}

func convdomainGroupToGroupRes(src domain.Group) (dst GroupRes) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.Description = src.Description
	dst.JoinFreely = src.JoinFreely
	dst.Members = make([]uuid.UUID, len(src.Members))
	for i := range src.Members {
		dst.Members[i] = convdomainUserTouuidUUID(src.Members[i])
	}
	dst.Admins = make([]uuid.UUID, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = convdomainUserTouuidUUID(src.Admins[i])
	}
	dst.IsTraQGroup = src.IsTraQGroup
	dst.CreatedBy = convdomainUserTouuidUUID(src.CreatedBy)
	dst.Model = Model(src.Model)
	return
}

func convdomainScheduleStatusToScheduleStatus(src domain.ScheduleStatus) (dst ScheduleStatus) {
	dst = ScheduleStatus(src)
	return
}

func convdomainStartEndTimeToStartEndTime(src domain.StartEndTime) (dst StartEndTime) {
	dst = StartEndTime(src)
	return
}

func convdomainTagToTagRes(src domain.Tag) (dst TagRes) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.Model = Model(src.Model)
	return
}

func convdomainTagTouuidUUID(src domain.Tag) (dst uuid.UUID) {
	dst = src.ID
	return
}

func convdomainUserToUserRes(src domain.User) (dst UserRes) {
	dst = UserRes(src)
	return
}

func convdomainUserTouuidUUID(src domain.User) (dst uuid.UUID) {
	dst = src.ID
	return
}
