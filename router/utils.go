package router

import (
	"net/url"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filters"
	"github.com/traPtitech/knoQ/router/presentation"
)

func getUserRelationFilter(values url.Values, userID uuid.UUID) filters.Expr {
	urel := presentation.GetUserRelationQuery(values)
	switch urel {
	case presentation.RelationBelongs:
		return filters.FilterBelongs(userID)
	case presentation.RelationAdmins:
		return filters.FilterAdmins(userID)
	}

	return filters.FilterBelongs(userID)
}

func createUserMap(users []*domain.User) map[uuid.UUID]*domain.User {
	userMap := make(map[uuid.UUID]*domain.User)
	for _, user := range users {
		userMap[user.ID] = user
	}
	return userMap
}
