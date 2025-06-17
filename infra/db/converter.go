package db

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

func ConvCreateRoomParamsToRoom(src CreateRoomParams) (dst Room) {
	dst.CreatedByRefer = src.CreatedBy
	dst.Name = src.Place
	dst.TimeStart = src.TimeStart
	dst.TimeEnd = src.TimeEnd
	dst.Admins = lo.Map(src.Admins, func(i uuid.UUID, _ int) *User {
		u := new(User)
		*u = ConvuuidUUIDToUserMeta(i)
		return u
	})
	return
}

func ConvEventAdminToRoomAdmin(src EventAdmin) (dst RoomAdmin) {
	dst.UserID = src.UserID
	return
}

func ConvEventAdminTodomainUser(src EventAdmin) (dst domain.User) {
	dst.ID = src.UserID
	return
}

func ConvEventAttendeeTodomainAttendee(src EventAttendee) (dst domain.Attendee) {
	dst.UserID = src.UserID
	dst.Schedule = domain.ScheduleStatus(src.Schedule)
	return
}

func ConvEventTagTodomainEventTag(src EventTag) (dst domain.EventTag) {
	dst.Tag = ConvTagTodomainTag(src.Tag)
	dst.Locked = src.Locked
	return
}

func ConvEventTodomainEvent(src Event) (dst domain.Event) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.Description = src.Description
	dst.IsRoomEvent = src.IsRoomEvent
	if dst.IsRoomEvent && src.Room != nil {
		dst.Room = new(domain.Room)
		*dst.Room = convRoomTodomainRoom(*src.Room)
	}
	dst.Venue = src.Venue
	dst.Group = convGroupTodomainGroup(src.Group)
	dst.TimeStart = src.TimeStart
	dst.TimeEnd = src.TimeEnd
	dst.CreatedBy = convUserTodomainUser(src.CreatedBy)
	dst.Admins = make([]domain.User, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = convEventAdminTodomainUser(src.Admins[i])
	}
	dst.Tags = make([]domain.EventTag, len(src.Tags))
	for i := range src.Tags {
		dst.Tags[i] = convEventTagTodomainEventTag(src.Tags[i])
	}
	dst.AllowTogether = src.AllowTogether
	dst.Attendees = make([]domain.Attendee, len(src.Attendees))
	for i := range src.Attendees {
		dst.Attendees[i] = convEventAttendeeTodomainAttendee(src.Attendees[i])
	}
	dst.Open = src.Open
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
	dst.DeletedAt = new(time.Time)
	(*dst.DeletedAt) = convgormDeletedAtTotimeTime(src.DeletedAt)
	return
}

func ConvGroupAdminTodomainUser(src GroupAdmin) (dst domain.User) {
	dst.ID = src.UserID
	return
}

func ConvGroupMemberTodomainUser(src GroupMember) (dst domain.User) {
	dst.ID = src.UserID
	return
}

func ConvGroupTodomainGroup(src Group) (dst domain.Group) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.Description = src.Description
	dst.JoinFreely = src.JoinFreely
	dst.Members = make([]domain.User, len(src.Members))
	for i := range src.Members {
		dst.Members[i] = convGroupMemberTodomainUser(src.Members[i])
	}
	dst.Admins = make([]domain.User, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = convGroupAdminTodomainUser(src.Admins[i])
	}
	dst.CreatedBy = convUserTodomainUser(src.CreatedBy)
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
	dst.DeletedAt = new(time.Time)
	(*dst.DeletedAt) = convgormDeletedAtTotimeTime(src.DeletedAt)
	return
}

func ConvRoomAdminTodomainUser(src RoomAdmin) (dst domain.User) {
	dst.ID = src.UserID
	return
}

func ConvRoomTodomainRoom(src Room) (dst domain.Room) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.TimeStart = src.TimeStart
	dst.TimeEnd = src.TimeEnd
	dst.Events = make([]domain.Event, len(src.Events))
	for i := range src.Events {
		dst.Events[i] = convEventTodomainEvent(*src.Events[i])
	}
	dst.Admins = make([]domain.User, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = convUserTodomainUser(*src.Admins[i])
	}
	dst.CreatedBy = convUserTodomainUser(src.CreatedBy)
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
	dst.DeletedAt = new(time.Time)
	(*dst.DeletedAt) = convgormDeletedAtTotimeTime(src.DeletedAt)
	return
}

func ConvSEventAdminToSRoomAdmin(src []EventAdmin) (dst []RoomAdmin) {
	dst = make([]RoomAdmin, len(src))
	for i := range src {
		dst[i] = convEventAdminToRoomAdmin(src[i])
	}
	return
}

func ConvSPEventToSPdomainEvent(src []*Event) []*domain.Event {
	return lo.Map(src, func(e *Event, _ int) *domain.Event {
		event := new(domain.Event)
		*event = convEventTodomainEvent((*e))
		return event
	})
}

func ConvSPGroupMemberToSuuidUUID(src []*GroupMember) (dst []uuid.UUID) {
	dst = make([]uuid.UUID, len(src))
	for i := range src {
		if src[i] != nil {
			dst[i] = src[i].GroupID
		}
	}
	return
}

func ConvSPGroupToSPdomainGroup(src []*Group) (dst []*domain.Group) {
	dst = make([]*domain.Group, len(src))
	for i := range src {
		if src[i] != nil {
			dst[i] = new(domain.Group)
			(*dst[i]) = convGroupTodomainGroup((*src[i]))
		}
	}
	return
}

func ConvSPRoomToSPdomainRoom(src []*Room) (dst []*domain.Room) {
	dst = make([]*domain.Room, len(src))
	for i := range src {
		if src[i] != nil {
			dst[i] = new(domain.Room)
			(*dst[i]) = convRoomTodomainRoom((*src[i]))
		}
	}
	return
}

func ConvSPTagToSPdomainTag(src []*Tag) (dst []*domain.Tag) {
	dst = make([]*domain.Tag, len(src))
	for i := range src {
		if src[i] != nil {
			dst[i] = new(domain.Tag)
			(*dst[i]) = convTagTodomainTag((*src[i]))
		}
	}
	return
}

func ConvTagTodomainEventTag(src Tag) (dst domain.EventTag) {
	dst.Tag.ID = src.ID
	dst.Tag.Name = src.Name
	return
}

func ConvTagTodomainTag(src Tag) (dst domain.Tag) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
	dst.DeletedAt = new(time.Time)
	(*dst.DeletedAt) = convgormDeletedAtTotimeTime(src.DeletedAt)
	return
}

func ConvUpdateRoomParamsToRoom(src UpdateRoomParams) (dst Room) {
	dst.CreatedByRefer = src.CreatedBy
	dst.Name = src.Place
	dst.TimeStart = src.TimeStart
	dst.TimeEnd = src.TimeEnd
	dst.Admins = make([]*User, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = new(User)
		*dst.Admins[i] = ConvuuidUUIDToUserMeta(src.Admins[i])
	}
	return
}

func ConvUserMetaTodomainUser(src User) (dst domain.User) {
	dst.ID = src.ID
	return
}

func ConvUserTodomainUser(src User) (dst domain.User) {
	dst.ID = src.ID
	dst.State = src.State
	return
}

func ConvWriteEventParamsToEvent(src WriteEventParams) (dst Event) {
	dst.CreatedByRefer = src.CreatedBy
	dst.Name = src.Name
	dst.Description = src.Description
	dst.GroupID = src.GroupID
	dst.IsRoomEvent = src.IsRoomEvent
	dst.RoomID = src.RoomID
	dst.Venue = src.Venue
	dst.TimeStart = src.TimeStart
	dst.TimeEnd = src.TimeEnd
	dst.Admins = make([]EventAdmin, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = convuuidUUIDToEventAdmin(src.Admins[i])
	}
	dst.AllowTogether = src.AllowTogether
	dst.Tags = make([]EventTag, len(src.Tags))
	for i := range src.Tags {
		dst.Tags[i] = convdomainEventTagParamsToEventTag(src.Tags[i])
	}
	dst.Open = src.Open
	return
}

func ConvWriteGroupParamsToGroup(src WriteGroupParams) (dst Group) {
	dst.CreatedByRefer = src.CreatedBy
	dst.Name = src.Name
	dst.Description = src.Description
	dst.JoinFreely = src.JoinFreely
	dst.Members = make([]GroupMember, len(src.Members))
	for i := range src.Members {
		dst.Members[i] = convuuidUUIDToGroupMember(src.Members[i])
	}
	dst.Admins = make([]GroupAdmin, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = convuuidUUIDToGroupAdmin(src.Admins[i])
	}
	return
}

func ConvdomainEventTagParamsToEventTag(src domain.EventTagParams) (dst EventTag) {
	dst.Tag.Name = src.Name
	dst.Locked = src.Locked
	return
}

func ConvgormDeletedAtTotimeTime(src gorm.DeletedAt) (dst time.Time) {
	dst = src.Time
	return
}

func ConvuuidUUIDToEventAdmin(src uuid.UUID) (dst EventAdmin) {
	dst.UserID = src
	return
}

func ConvuuidUUIDToGroupAdmin(src uuid.UUID) (dst GroupAdmin) {
	dst.UserID = src
	return
}

func ConvuuidUUIDToGroupAdmins(src uuid.UUID) (dst GroupAdmin) {
	dst.UserID = src
	return
}

func ConvuuidUUIDToGroupMember(src uuid.UUID) (dst GroupMember) {
	dst.UserID = src
	return
}

func ConvuuidUUIDToRoomAdmin(src uuid.UUID) (dst RoomAdmin) {
	dst.UserID = src
	return
}

func ConvuuidUUIDToUserMeta(src uuid.UUID) (dst User) {
	dst.ID = src
	return
}

func convEventAdminToRoomAdmin(src EventAdmin) (dst RoomAdmin) {
	dst.UserID = src.UserID
	return
}

func convEventAdminTodomainUser(src EventAdmin) (dst domain.User) {
	dst.ID = src.UserID
	return
}

func convEventAttendeeTodomainAttendee(src EventAttendee) (dst domain.Attendee) {
	dst.UserID = src.UserID
	dst.Schedule = domain.ScheduleStatus(src.Schedule)
	return
}

func convEventTagTodomainEventTag(src EventTag) (dst domain.EventTag) {
	dst.Tag = convTagTodomainTag(src.Tag)
	dst.Locked = src.Locked
	return
}

func convEventTodomainEvent(src Event) (dst domain.Event) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.IsRoomEvent = src.IsRoomEvent
	if dst.IsRoomEvent && src.Room != nil {
		dst.Room = new(domain.Room)
		*dst.Room = convRoomTodomainRoom(*src.Room)
	} else {
		dst.Venue = src.Venue
	}
	dst.Description = src.Description
	dst.Group = convGroupTodomainGroup(src.Group)
	dst.Group.ID = src.GroupID
	dst.TimeStart = src.TimeStart
	dst.TimeEnd = src.TimeEnd
	dst.CreatedBy = convUserTodomainUser(src.CreatedBy)
	dst.Admins = make([]domain.User, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = convEventAdminTodomainUser(src.Admins[i])
	}
	dst.Tags = make([]domain.EventTag, len(src.Tags))
	for i := range src.Tags {
		dst.Tags[i] = convEventTagTodomainEventTag(src.Tags[i])
	}
	dst.AllowTogether = src.AllowTogether
	dst.Attendees = make([]domain.Attendee, len(src.Attendees))
	for i := range src.Attendees {
		dst.Attendees[i] = convEventAttendeeTodomainAttendee(src.Attendees[i])
	}
	dst.Open = src.Open
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
	dst.DeletedAt = new(time.Time)
	(*dst.DeletedAt) = convgormDeletedAtTotimeTime(src.DeletedAt)
	return
}

func convGroupAdminTodomainUser(src GroupAdmin) (dst domain.User) {
	dst.ID = src.UserID
	return
}

func convGroupMemberTodomainUser(src GroupMember) (dst domain.User) {
	dst.ID = src.UserID
	return
}

func convGroupTodomainGroup(src Group) (dst domain.Group) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.Description = src.Description
	dst.JoinFreely = src.JoinFreely
	dst.Members = make([]domain.User, len(src.Members))
	for i := range src.Members {
		dst.Members[i] = convGroupMemberTodomainUser(src.Members[i])
	}
	dst.Admins = make([]domain.User, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = convGroupAdminTodomainUser(src.Admins[i])
	}
	dst.CreatedBy = convUserTodomainUser(src.CreatedBy)
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
	dst.DeletedAt = new(time.Time)
	(*dst.DeletedAt) = convgormDeletedAtTotimeTime(src.DeletedAt)
	return
}

func convRoomTodomainRoom(src Room) (dst domain.Room) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.TimeStart = src.TimeStart
	dst.TimeEnd = src.TimeEnd
	dst.Events = make([]domain.Event, len(src.Events))
	for i := range src.Events {
		dst.Events[i] = convEventTodomainEvent(*src.Events[i])
	}
	dst.Admins = make([]domain.User, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = convUserTodomainUser(*src.Admins[i])
	}
	dst.CreatedBy = convUserTodomainUser(src.CreatedBy)
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
	dst.DeletedAt = new(time.Time)
	(*dst.DeletedAt) = convgormDeletedAtTotimeTime(src.DeletedAt)
	return
}

func convTagTodomainTag(src Tag) (dst domain.Tag) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
	dst.DeletedAt = new(time.Time)
	(*dst.DeletedAt) = convgormDeletedAtTotimeTime(src.DeletedAt)
	return
}

func convUserTodomainUser(src User) (dst domain.User) {
	dst.ID = src.ID
	dst.State = src.State
	return
}

func convdomainEventTagParamsToEventTag(src domain.EventTagParams) (dst EventTag) {
	dst.Tag.Name = src.Name
	dst.Locked = src.Locked
	return
}

func convgormDeletedAtTotimeTime(src gorm.DeletedAt) (dst time.Time) {
	dst = src.Time
	return
}

func convuuidUUIDToEventAdmin(src uuid.UUID) (dst EventAdmin) {
	dst.UserID = src
	return
}

func convuuidUUIDToGroupAdmin(src uuid.UUID) (dst GroupAdmin) {
	dst.UserID = src
	return
}

func convuuidUUIDToGroupMember(src uuid.UUID) (dst GroupMember) {
	dst.UserID = src
	return
}
