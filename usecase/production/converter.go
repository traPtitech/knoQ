// Code generated by gotypeconverter; DO NOT EDIT.
package production

import (
	"github.com/traPtitech/knoQ/domain"
	v3 "github.com/traPtitech/traQ/router/v3"
)

func ConvertPointerv3UserToPointerdomainUser(src *v3.User) (dst *domain.User) {
	dst = new(domain.User)
	(*dst) = Convertv3UserTodomainUser((*src))
	return
}

func Convertv3UserTodomainUser(src v3.User) (dst domain.User) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.DisplayName = src.DisplayName
	return
}