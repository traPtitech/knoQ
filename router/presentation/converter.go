package presentation

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
)

func ConvEventReqWriteTodomainWriteEventParams(src EventReqWrite) (dst domain.WriteEventParams) {
	dst.Name = src.Name
	dst.Description = src.Description
	dst.GroupID = src.GroupID
	dst.RoomID = src.RoomID
	dst.Place = src.Place
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

func ConvSPdomainEventToSEventRes(src []*domain.Event) (dst []EventRes) {
	dst = make([]EventRes, len(src))
	for i := range src {
		if src[i] != nil {
			dst[i].ID = src[i].ID
			dst[i].Name = src[i].Name
			dst[i].Description = src[i].Description
			dst[i].AllowTogether = src[i].AllowTogether
			dst[i].TimeStart = src[i].TimeStart
			dst[i].TimeEnd = src[i].TimeEnd
			dst[i].RoomID = convdomainRoomTouuidUUID(src[i].Room)
			dst[i].GroupID = convdomainGroupTouuidUUID(src[i].Group)
			dst[i].Place = src[i].Room.Place
			dst[i].GroupName = src[i].Group.Name
			dst[i].Admins = make([]uuid.UUID, len(src[i].Admins))
			for j := range src[i].Admins {
				dst[i].Admins[j] = convdomainUserTouuidUUID(src[i].Admins[j])
			}
			dst[i].Tags = make([]EventTagRes, len(src[i].Tags))
			for j := range src[i].Tags {
				dst[i].Tags[j] = convdomainEventTagToEventTagRes(src[i].Tags[j])
			}
			dst[i].CreatedBy = convdomainUserTouuidUUID(src[i].CreatedBy)
			dst[i].Open = src[i].Open
			dst[i].Attendees = make([]EventAttendeeRes, len(src[i].Attendees))
			for j := range src[i].Attendees {
				dst[i].Attendees[j] = convdomainAttendeeToEventAttendeeRes(src[i].Attendees[j])
			}
			dst[i].Model = Model(src[i].Model)
		}
	}
	return
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

func ConvdomainEventToEventRes(src domain.Event) (dst EventRes) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.Description = src.Description
	dst.AllowTogether = src.AllowTogether
	dst.TimeStart = src.TimeStart
	dst.TimeEnd = src.TimeEnd
	dst.RoomID = convdomainRoomTouuidUUID(src.Room)
	dst.GroupID = convdomainGroupTouuidUUID(src.Group)
	dst.Place = src.Room.Place
	dst.GroupName = src.Group.Name
	dst.Admins = make([]uuid.UUID, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = convdomainUserTouuidUUID(src.Admins[i])
	}
	dst.Tags = make([]EventTagRes, len(src.Tags))
	for i := range src.Tags {
		dst.Tags[i] = convdomainEventTagToEventTagRes(src.Tags[i])
	}
	dst.CreatedBy = convdomainUserTouuidUUID(src.CreatedBy)
	dst.Open = src.Open
	dst.Attendees = make([]EventAttendeeRes, len(src.Attendees))
	for i := range src.Attendees {
		dst.Attendees[i] = convdomainAttendeeToEventAttendeeRes(src.Attendees[i])
	}
	dst.Model = Model(src.Model)
	return
}

func ConvdomainGroupToGroupRes(src domain.Group) (dst GroupRes) {
	dst.ID = src.ID
	dst.GroupReq.Name = src.Name
	dst.GroupReq.Description = src.Description
	dst.GroupReq.JoinFreely = src.JoinFreely
	dst.GroupReq.Members = make([]uuid.UUID, len(src.Members))
	for i := range src.Members {
		dst.GroupReq.Members[i] = convdomainUserTouuidUUID(src.Members[i])
	}
	dst.GroupReq.Admins = make([]uuid.UUID, len(src.Admins))
	for i := range src.Admins {
		dst.GroupReq.Admins[i] = convdomainUserTouuidUUID(src.Admins[i])
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
	dst.GroupReq.Name = src.Name
	dst.GroupReq.Description = src.Description
	dst.GroupReq.JoinFreely = src.JoinFreely
	dst.GroupReq.Members = make([]uuid.UUID, len(src.Members))
	for i := range src.Members {
		dst.GroupReq.Members[i] = convdomainUserTouuidUUID(src.Members[i])
	}
	dst.GroupReq.Admins = make([]uuid.UUID, len(src.Admins))
	for i := range src.Admins {
		dst.GroupReq.Admins[i] = convdomainUserTouuidUUID(src.Admins[i])
	}
	dst.IsTraQGroup = src.IsTraQGroup
	dst.CreatedBy = convdomainUserTouuidUUID(src.CreatedBy)
	dst.Model = Model(src.Model)
	return
}

func convdomainGroupTouuidUUID(src domain.Group) (dst uuid.UUID) {
	dst = src.ID
	return
}

func convdomainRoomTouuidUUID(src domain.Room) (dst uuid.UUID) {
	dst = src.ID
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
