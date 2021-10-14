package router

import (
	"net/url"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filter"
	"github.com/traPtitech/knoQ/presentation"
)

func getUserRelationFilter(values url.Values, userID uuid.UUID) filter.Expr {
	urel := presentation.GetUserRelationQuery(values)
	switch urel {
	case presentation.RelationBelongs:
		return filter.FilterBelongs(userID)
	case presentation.RelationAdmins:
		return filter.FilterAdmins(userID)
	}

	return filter.FilterBelongs(userID)
}

func createUserMap(users []*domain.User) map[uuid.UUID]*domain.User {
	userMap := make(map[uuid.UUID]*domain.User)
	for _, user := range users {
		userMap[user.ID] = user
	}
	return userMap
}
