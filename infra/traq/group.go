package traq

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"

	"github.com/traPtitech/go-traq"
	"github.com/traPtitech/knoQ/infra"
)

func (repo *traqRepository) GetGroup(groupID uuid.UUID) (*infra.TraqUserGroupResponse, error) {
	ctx := context.WithValue(context.TODO(), traq.ContextAccessToken, repo.ServerAccessToken)
	apiClient := traq.NewAPIClient(traq.NewConfiguration())
	// TODO: 一定期間キャッシュする
	g, resp, err := apiClient.GroupApi.GetUserGroup(ctx, groupID.String()).Execute()
	if err != nil {
		return nil, err
	}
	err = handleStatusCode(resp.StatusCode)
	if err != nil {
		return nil, err
	}

	group := infra.TraqUserGroupResponse{
		ID:          uuid.FromStringOrNil(g.Id),
		Name:        g.Name,
		Description: g.Description,
		Type:        g.Type,
		Members: lo.Map(g.Members, func(m traq.UserGroupMember, _ int) infra.TraqUserGroupMember {
			return infra.TraqUserGroupMember{
				ID:   uuid.FromStringOrNil(m.Id),
				Role: m.Role,
			}
		}),
		IconID:    uuid.FromStringOrNil(g.Icon),
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
		Admins: lo.Map(g.Admins, func(adminID string, _ int) uuid.UUID {
			return uuid.FromStringOrNil(adminID)
		}),
	}

	return &group, err
}

func (repo *traqRepository) GetAllGroups() ([]*infra.TraqUserGroupResponse, error) {
	ctx := context.WithValue(context.TODO(), traq.ContextAccessToken, repo.ServerAccessToken)
	apiClient := traq.NewAPIClient(traq.NewConfiguration())
	// TODO: 一定期間キャッシュする
	gs, resp, err := apiClient.GroupApi.GetUserGroups(ctx).Execute()
	if err != nil {
		return nil, err
	}
	err = handleStatusCode(resp.StatusCode)
	if err != nil {
		return nil, err
	}

	groups := lo.Map(gs, func(g traq.UserGroup, _ int) *infra.TraqUserGroupResponse {
		return &infra.TraqUserGroupResponse{
			ID:          uuid.FromStringOrNil(g.Id),
			Name:        g.Name,
			Description: g.Description,
			Type:        g.Type,
			Members: lo.Map(g.Members, func(m traq.UserGroupMember, _ int) infra.TraqUserGroupMember {
				return infra.TraqUserGroupMember{
					ID:   uuid.FromStringOrNil(m.Id),
					Role: m.Role,
				}
			}),
			IconID:    uuid.FromStringOrNil(g.Icon),
			CreatedAt: g.CreatedAt,
			UpdatedAt: g.UpdatedAt,
			Admins: lo.Map(g.Admins, func(adminID string, _ int) uuid.UUID {
				return uuid.FromStringOrNil(adminID)
			}),
		}
	})

	return groups, err
}

// これ accessToken いるのなんで?
func (repo *traqRepository) GetUserBelongingGroupIDs(accessToken string, userID uuid.UUID) ([]uuid.UUID, error) {
	ctx := context.WithValue(context.TODO(), traq.ContextAccessToken, accessToken)
	apiClient := traq.NewAPIClient(traq.NewConfiguration())
	user, resp, err := apiClient.UserApi.GetUser(ctx, userID.String()).Execute()
	if err != nil {
		return nil, err
	}
	err = handleStatusCode(resp.StatusCode)
	if err != nil {
		return nil, err
	}
	groups := make([]uuid.UUID, 0, len(user.Groups))
	for _, gid := range user.Groups {
		groupUUID, err := uuid.FromString(gid)
		if err != nil {
			return nil, err
		}
		groups = append(groups, groupUUID)
	}
	return groups, err
}

func (repo *traqRepository) GetGradeGroups() ([]*infra.TraqUserGroupResponse, error) {
	ctx := context.WithValue(context.TODO(), traq.ContextAccessToken, repo.ServerAccessToken)
	apiClient := traq.NewAPIClient(traq.NewConfiguration())
	gs, resp, err := apiClient.GroupApi.GetUserGroups(ctx).Execute()
	if err != nil {
		return nil, err
	}
	err = handleStatusCode(resp.StatusCode)
	if err != nil {
		return nil, err
	}

	groups := lo.FilterMap(gs, func(g traq.UserGroup, _ int) (*infra.TraqUserGroupResponse, bool) {
		if g.Type != "grade" {
			return nil, false
		}
		return &infra.TraqUserGroupResponse{
			ID:          uuid.FromStringOrNil(g.Id),
			Name:        g.Name,
			Description: g.Description,
			Type:        g.Type,
			Members: lo.Map(g.Members, func(m traq.UserGroupMember, _ int) infra.TraqUserGroupMember {
				return infra.TraqUserGroupMember{
					ID:   uuid.FromStringOrNil(m.Id),
					Role: m.Role,
				}
			}),
			IconID:    uuid.FromStringOrNil(g.Icon),
			CreatedAt: g.CreatedAt,
			UpdatedAt: g.UpdatedAt,
			Admins: lo.Map(g.Admins, func(adminID string, _ int) uuid.UUID {
				return uuid.FromStringOrNil(adminID)
			}),
		}, true
	})

	return groups, err
}
