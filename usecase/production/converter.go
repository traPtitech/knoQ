// FIXME: ↓が動かないので一時的に手動で作成
// //go:generate gotypeconverter -s []*traq.UserGroup -d []*domain.Group -o converter.go .
package production

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/go-traq"
	"github.com/traPtitech/knoQ/domain"
)

func ConvPtraqUserToPdomainUser(src *traq.User) (dst *domain.User) {
	dst = new(domain.User)
	(*dst) = ConvtraqUserTodomainUser((*src))
	return
}

func ConvSPtraqUserGroupToSPdomainGroup(src []*traq.UserGroup) (dst []*domain.Group) {
	dst = make([]*domain.Group, len(src))
	for i := range src {
		if src[i] != nil {
			dst[i] = new(domain.Group)
			(*dst[i]) = ConvtraqUserGroupTodomainGroup((*src[i]))
		}
	}
	return
}

func ConvuuidUUIDTodomainUser(src uuid.UUID) (dst domain.User) {
	dst.ID = src
	return
}
func ConvtraqUserGroupMemberTodomainUser(src traq.UserGroupMember) (dst domain.User) {
	dst.ID = uuid.Must(uuid.FromString(src.GetId()))
	return
}
func ConvtraqUserGroupTodomainGroup(src traq.UserGroup) (dst domain.Group) {
	dst.ID = uuid.Must(uuid.FromString(src.GetId()))
	dst.Name = src.Name
	dst.Description = src.Description
	dst.Members = make([]domain.User, len(src.Members))
	for i := range src.Members {
		dst.Members[i] = ConvtraqUserGroupMemberTodomainUser(src.Members[i])
	}
	dst.Admins = make([]domain.User, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = ConvuuidUUIDTodomainUser(uuid.Must(uuid.FromString(src.GetAdmins()[i])))
	}
	dst.Model.CreatedAt = src.CreatedAt
	dst.Model.UpdatedAt = src.UpdatedAt
	return
}
func ConvtraqUserTodomainUser(src traq.User) (dst domain.User) {
	dst.ID = uuid.Must(uuid.FromString(src.GetId()))
	dst.Name = src.Name
	dst.DisplayName = src.DisplayName
	return
}
