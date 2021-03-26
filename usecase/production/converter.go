// Code generated by gotypeconverter; DO NOT EDIT.
package production

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	v3 "github.com/traPtitech/traQ/router/v3"
)

func ConvertPointerv3UserToPointerdomainUser(src *v3.User) (dst *domain.User) {
	dst = new(domain.User)
	(*dst) = Convertv3UserTodomainUser((*src))
	return
}

func ConvertuuidUUIDTodomainUser(src uuid.UUID) (dst domain.User) {
	dst.ID = src
	return
}
func Convertv3UserGroupMemberTodomainUser(src v3.UserGroupMember) (dst domain.User) {
	dst.ID = src.ID
	return
}
func Convertv3UserGroupTodomainGroup(src v3.UserGroup) (dst domain.Group) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.Description = src.Description
	dst.Members = make([]domain.User, len(src.Members))
	for i := range src.Members {
		dst.Members[i] = Convertv3UserGroupMemberTodomainUser(src.Members[i])
	}
	dst.Admins = make([]domain.User, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = ConvertuuidUUIDTodomainUser(src.Admins[i])
	}
	dst.Model.CreatedAt = src.CreatedAt
	dst.Model.UpdatedAt = src.UpdatedAt
	return
}
func Convertv3UserTodomainUser(src v3.User) (dst domain.User) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.DisplayName = src.DisplayName
	return
}
