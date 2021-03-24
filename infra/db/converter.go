// Code generated by gotypeconverter; DO NOT EDIT.
package db

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

func ConvertEventAdminTodomainUser(src EventAdmin) (dst domain.User) {
	dst.ID = src.UserID
	return
}

func ConvertEventTagTodomainEventTag(src EventTag) (dst domain.EventTag) {
	dst.Tag = ConvertTagTodomainTag(src.Tag)
	dst.Locked = src.Locked
	return
}
func ConvertEventTodomainEvent(src Event) (dst domain.Event) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.Description = src.Description
	dst.Room = ConvertRoomTodomainRoom(src.Room)
	dst.Group = ConvertGroupTodomainGroup(src.Group)
	dst.TimeStart = src.TimeStart
	dst.TimeEnd = src.TimeEnd
	dst.CreatedBy = ConvertUserTodomainUser(src.CreatedBy)
	dst.Admins = make([]domain.User, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = ConvertEventAdminTodomainUser(src.Admins[i])
	}
	dst.Tags = make([]domain.EventTag, len(src.Tags))
	for i := range src.Tags {
		dst.Tags[i] = ConvertEventTagTodomainEventTag(src.Tags[i])
	}
	dst.AllowTogether = src.AllowTogether
	dst.Model.CreatedAt = src.Model.CreatedAt
	dst.Model.UpdatedAt = src.Model.UpdatedAt
	dst.Model.DeletedAt = new(time.Time)
	(*dst.Model.DeletedAt) = ConvertgormDeletedAtTotimeTime(src.Model.DeletedAt)
	return
}

func ConvertGroupAdminTodomainUser(src GroupAdmin) (dst domain.User) {
	dst.ID = src.UserID
	return
}
func ConvertGroupAdminsTodomainUser(src GroupAdmin) (dst domain.User) {
	dst.ID = src.UserID
	return
}

func ConvertGroupMemberTodomainUser(src GroupMember) (dst domain.User) {
	dst.ID = src.UserID
	return
}
func ConvertGroupTodomainGroup(src Group) (dst domain.Group) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.Description = src.Description
	dst.JoinFreely = src.JoinFreely
	dst.Members = make([]domain.User, len(src.Members))
	for i := range src.Members {
		dst.Members[i] = ConvertGroupMemberTodomainUser(src.Members[i])
	}
	dst.Admins = make([]domain.User, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = ConvertGroupAdminTodomainUser(src.Admins[i])
	}
	dst.CreatedBy = ConvertUserTodomainUser(src.CreatedBy)
	dst.Model.CreatedAt = src.Model.CreatedAt
	dst.Model.UpdatedAt = src.Model.UpdatedAt
	dst.Model.DeletedAt = new(time.Time)
	(*dst.Model.DeletedAt) = ConvertgormDeletedAtTotimeTime(src.Model.DeletedAt)
	dst.IsTraQGroup = src.Model.DeletedAt.Valid
	return
}
func ConvertRoomTodomainRoom(src Room) (dst domain.Room) {
	dst.ID = src.ID
	dst.Place = src.Place
	dst.Verified = src.Verified
	dst.TimeStart = src.TimeStart
	dst.TimeEnd = src.TimeEnd
	dst.Events = make([]domain.Event, len(src.Events))
	for i := range src.Events {
		dst.Events[i] = ConvertEventTodomainEvent(src.Events[i])
	}
	dst.CreatedBy = ConvertUserTodomainUser(src.CreatedBy)
	dst.Model.CreatedAt = src.Model.CreatedAt
	dst.Model.UpdatedAt = src.Model.UpdatedAt
	dst.Model.DeletedAt = new(time.Time)
	(*dst.Model.DeletedAt) = ConvertgormDeletedAtTotimeTime(src.Model.DeletedAt)
	return
}
func ConvertSlicePointerGroupMemberToSliceuuidUUID(src []*GroupMember) (dst []uuid.UUID) {
	dst = make([]uuid.UUID, len(src))
	for i := range src {
		dst[i] = (*src[i]).GroupID
	}
	return
}
func ConvertTagTodomainEventTag(src Tag) (dst domain.EventTag) {
	dst.Tag.ID = src.ID
	dst.Tag.Name = src.Name
	return
}

func ConvertTagTodomainTag(src Tag) (dst domain.Tag) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.Model.CreatedAt = src.Model.CreatedAt
	dst.Model.UpdatedAt = src.Model.UpdatedAt
	dst.Model.DeletedAt = new(time.Time)
	(*dst.Model.DeletedAt) = ConvertgormDeletedAtTotimeTime(src.Model.DeletedAt)
	return
}
func ConvertUserMetaTodomainUser(src User) (dst domain.User) {
	dst.ID = src.ID
	return
}

func ConvertUserTodomainUser(src User) (dst domain.User) {
	dst.ID = src.ID
	return
}
func ConvertWriteEventParamsToEvent(src WriteEventParams) (dst Event) {
	dst.CreatedByRefer = src.CreatedBy
	dst.Name = src.WriteEventParams.Name
	dst.Description = src.WriteEventParams.Description
	dst.GroupID = src.WriteEventParams.GroupID
	dst.RoomID = src.WriteEventParams.RoomID
	dst.TimeStart = src.WriteEventParams.TimeStart
	dst.TimeEnd = src.WriteEventParams.TimeEnd
	dst.Admins = make([]EventAdmin, len(src.WriteEventParams.Admins))
	for i := range src.WriteEventParams.Admins {
		dst.Admins[i] = ConvertuuidUUIDToEventAdmin(src.WriteEventParams.Admins[i])
	}
	dst.AllowTogether = src.WriteEventParams.AllowTogether
	dst.Tags = make([]EventTag, len(src.WriteEventParams.Tags))
	for i := range src.WriteEventParams.Tags {
		dst.Tags[i] = ConvertdomainEventTagParamsToEventTag(src.WriteEventParams.Tags[i])
	}
	return
}

func ConvertdomainEventTagParamsToEventTag(src domain.EventTagParams) (dst EventTag) {
	dst.Tag.Name = src.Name
	dst.Locked = src.Locked
	return
}

func ConvertgormDeletedAtTotimeTime(src gorm.DeletedAt) (dst time.Time) {
	dst = src.Time
	return
}
func ConvertuuidUUIDToEventAdmin(src uuid.UUID) (dst EventAdmin) {
	dst.UserID = src
	return
}
func ConvertuuidUUIDToGroupAdmin(src uuid.UUID) (dst GroupAdmin) {
	dst.UserID = src
	return
}
func ConvertuuidUUIDToGroupAdmins(src uuid.UUID) (dst GroupAdmin) {
	dst.UserID = src
	return
}

func ConvertuuidUUIDToGroupMember(src uuid.UUID) (dst GroupMember) {
	dst.UserID = src
	return
}
func ConvertuuidUUIDToUserMeta(src uuid.UUID) (dst User) {
	dst.ID = src
	return
}

func ConvertwriteEventParamsToEvent(src WriteEventParams) (dst Event) {
	dst.CreatedByRefer = src.CreatedBy
	dst.Name = src.WriteEventParams.Name
	dst.Description = src.WriteEventParams.Description
	dst.GroupID = src.WriteEventParams.GroupID
	dst.RoomID = src.WriteEventParams.RoomID
	dst.TimeStart = src.WriteEventParams.TimeStart
	dst.TimeEnd = src.WriteEventParams.TimeEnd
	dst.Admins = make([]EventAdmin, len(src.WriteEventParams.Admins))
	for i := range src.WriteEventParams.Admins {
		dst.Admins[i] = ConvertuuidUUIDToEventAdmin(src.WriteEventParams.Admins[i])
	}
	dst.AllowTogether = src.WriteEventParams.AllowTogether
	dst.Tags = make([]EventTag, len(src.WriteEventParams.Tags))
	for i := range src.WriteEventParams.Tags {
		dst.Tags[i] = ConvertdomainEventTagParamsToEventTag(src.WriteEventParams.Tags[i])
	}
	return
}

func ConvertwriteGroupParamsToGroup(src writeGroupParams) (dst Group) {
	dst.CreatedByRefer = src.CreatedBy
	dst.Name = src.WriteGroupParams.Name
	dst.Description = src.WriteGroupParams.Description
	dst.JoinFreely = src.WriteGroupParams.JoinFreely
	dst.Members = make([]GroupMember, len(src.WriteGroupParams.Members))
	for i := range src.WriteGroupParams.Members {
		dst.Members[i] = ConvertuuidUUIDToGroupMember(src.WriteGroupParams.Members[i])
	}
	dst.Admins = make([]GroupAdmin, len(src.WriteGroupParams.Admins))
	for i := range src.WriteGroupParams.Admins {
		dst.Admins[i] = ConvertuuidUUIDToGroupAdmin(src.WriteGroupParams.Admins[i])
	}
	return
}
func ConvertwriteRoomParamsToRoom(src writeRoomParams) (dst Room) {
	dst.Verified = src.Verified
	dst.CreatedByRefer = src.CreatedBy
	dst.Place = src.WriteRoomParams.Place
	dst.TimeStart = src.WriteRoomParams.TimeStart
	dst.TimeEnd = src.WriteRoomParams.TimeEnd
	return
}
