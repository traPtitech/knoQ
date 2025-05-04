package repository

import (
	"errors"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra"
	"github.com/traPtitech/knoQ/infra/db"
	"gorm.io/gorm"
)

var traPGroupID = uuid.Must(uuid.FromString("11111111-1111-1111-1111-111111111111"))

func (repo *Repository) CreateGroup(params domain.WriteGroupParams, info *domain.ConInfo) (*domain.Group, error) {
	p := db.WriteGroupParams{
		WriteGroupParams: params,
		CreatedBy:        info.ReqUserID,
	}
	g, err := repo.GormRepo.CreateGroup(p)
	if err != nil {
		return nil, err
	}
	// group := db.ConvGroupTodomainGroup(*g)
	return g, nil
}

func (repo *Repository) UpdateGroup(groupID uuid.UUID, params domain.WriteGroupParams, info *domain.ConInfo) (*domain.Group, error) {
	if !repo.IsGroupAdmins(groupID, info) {
		return nil, domain.ErrForbidden
	}
	p := db.WriteGroupParams{
		WriteGroupParams: params,
		CreatedBy:        info.ReqUserID,
	}
	g, err := repo.GormRepo.UpdateGroup(groupID, p)
	if err != nil {
		return nil, err
	}
	// group := db.ConvGroupTodomainGroup(*g)
	return g, nil
}

// AddMeToGroup add me to that group if that group is open.
func (repo *Repository) AddMeToGroup(groupID uuid.UUID, info *domain.ConInfo) error {
	if !repo.IsGroupJoinFreely(groupID) {
		return domain.ErrForbidden
	}
	return repo.GormRepo.AddMemberToGroup(groupID, info.ReqUserID)
}

func (repo *Repository) DeleteGroup(groupID uuid.UUID, info *domain.ConInfo) error {
	if !repo.IsGroupAdmins(groupID, info) {
		return domain.ErrForbidden
	}
	return repo.GormRepo.DeleteGroup(groupID)
}

// DeleteMeGroup delete me in that group if that group is open.
func (repo *Repository) DeleteMeGroup(groupID uuid.UUID, info *domain.ConInfo) error {
	if !repo.IsGroupJoinFreely(groupID) {
		return domain.ErrForbidden
	}
	return repo.GormRepo.DeleteMemberOfGroup(groupID, info.ReqUserID)
}

func (repo *Repository) GetGroup(groupID uuid.UUID, info *domain.ConInfo) (*domain.Group, error) {
	var group *domain.Group
	g, err := repo.GormRepo.GetGroup(groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// trap
			if groupID == traPGroupID {
				return repo.getTraPGroup(info), nil
			}

			// traq group
			g, err := repo.GormRepo.GetTraqRepository().GetGroup(groupID)
			if err != nil {
				return nil, err
			}
			// group = ConvtraqUserGroupTodomainGroup(*g)
			// group.IsTraQGroup = true
			group = &domain.Group{
				ID:          g.ID,
				Name:        g.Name,
				Description: g.Description,
				Members: lo.Map(g.Members, func(m infra.TraqUserGroupMember, _ int) domain.User {
					return domain.User{ID: m.ID}
				}),
				Admins: lo.Map(g.Admins, func(adminID uuid.UUID, _ int) domain.User {
					return domain.User{ID: adminID}
				}),
				IsTraQGroup: true,
				Model: domain.Model{
					CreatedAt: g.CreatedAt,
					UpdatedAt: g.UpdatedAt,
				},
			}
		} else {
			return nil, err
		}
	} else {
		group = g
	}
	return group, nil
}

func (repo *Repository) GetAllGroups(info *domain.ConInfo) ([]*domain.Group, error) {
	groups := make([]*domain.Group, 0)
	gg, err := repo.GormRepo.GetAllGroups()
	if err != nil {
		return nil, err
	}
	groups = append(groups, gg...)
	tg, err := repo.GormRepo.GetTraqRepository().GetAllGroups()
	if err != nil {
		return nil, err
	}
	// dg := ConvSPtraqUserGroupToSPdomainGroup(tg)
	dg := lo.Map(tg, func(g *infra.TraqUserGroupResponse, _ int) *domain.Group {
		return &domain.Group{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Members: lo.Map(g.Members, func(m infra.TraqUserGroupMember, _ int) domain.User {
				return domain.User{ID: m.ID}
			}),
			Admins: lo.Map(g.Admins, func(adminID uuid.UUID, _ int) domain.User {
				return domain.User{ID: adminID}
			}),
			IsTraQGroup: true,
			Model: domain.Model{
				CreatedAt: g.CreatedAt,
				UpdatedAt: g.UpdatedAt,
			},
		}
	})
	// add IsTraQGroup
	// for i := range dg {
	// 	dg[i].IsTraQGroup = true
	// }
	// add trap
	groups = append(append(groups, repo.getTraPGroup(info)), dg...)

	return groups, nil
}

func (repo *Repository) GetUserBelongingGroupIDs(userID uuid.UUID, info *domain.ConInfo) ([]uuid.UUID, error) {
	t, err := repo.GormRepo.GetToken(info.ReqUserID)
	if err != nil {
		return nil, err
	}
	ggIDs, err := repo.GormRepo.GetBelongGroupIDs(userID)
	if err != nil {
		return nil, err
	}
	tgIDs, err := repo.GormRepo.GetTraqRepository().GetUserBelongingGroupIDs(t.AccessToken, userID)
	if err != nil {
		return nil, err
	}
	// add trap
	return append(append(ggIDs, traPGroupID), tgIDs...), nil
}

func (repo *Repository) GetUserAdminGroupIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	return repo.GormRepo.GetAdminGroupIDs(userID)
}

func (repo *Repository) IsGroupAdmins(groupID uuid.UUID, info *domain.ConInfo) bool {
	group, err := repo.GormRepo.GetGroup(groupID)
	if err != nil {
		return false
	}
	for _, admin := range group.Admins {
		if info.ReqUserID == admin.ID {
			return true
		}
	}
	return false
}

func (repo *Repository) IsGroupJoinFreely(groupID uuid.UUID) bool {
	group, err := repo.GormRepo.GetGroup(groupID)
	if err != nil {
		return false
	}
	return group.JoinFreely
}

func (repo *Repository) IsGroupMember(userID, groupID uuid.UUID, info *domain.ConInfo) bool {
	group, err := repo.GetGroup(groupID, info)
	if err != nil {
		return false
	}
	for _, member := range group.Members {
		if userID == member.ID {
			return true
		}
	}

	return false
}

func (repo *Repository) getTraPGroup(info *domain.ConInfo) *domain.Group {
	members, err := repo.GetAllUsers(false, false, info)
	if err != nil {
		return nil
	}

	return &domain.Group{
		ID:          traPGroupID,
		Name:        "traP",
		Description: "traP全体グループ",
		JoinFreely:  false,
		Members:     convSPdomainUserToSdomainUser(members),
		Admins:      []domain.User{},
		IsTraQGroup: true,
		CreatedBy:   domain.User{},
		Model:       domain.Model{},
	}
}

func (repo *Repository) GetGradeGroupNames(_ *domain.ConInfo) ([]string, error) {
	groups, err := repo.GormRepo.GetTraqRepository().GetAllGroups()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, g := range groups {
		if g.Type == "grade" {
			names = append(names, g.Name)
		}
	}

	return names, nil
}

func convSPdomainUserToSdomainUser(src []*domain.User) (dst []domain.User) {
	dst = make([]domain.User, len(src))
	for i := range src {
		if src[i] != nil {
			dst[i] = (*src[i])
		}
	}
	return
}
