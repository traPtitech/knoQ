package repository

import (
	"errors"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
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
		return nil, defaultErrorHandling(err)
	}
	group := db.ConvGroupTodomainGroup(*g)
	return &group, nil
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
		return nil, defaultErrorHandling(err)
	}
	group := db.ConvGroupTodomainGroup(*g)
	return &group, nil
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
	var group domain.Group
	g, err := repo.GormRepo.GetGroup(groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// trap
			if groupID == traPGroupID {
				return repo.getTraPGroup(info), nil
			}

			// traq group
			g, err := repo.TraQRepo.GetGroup(groupID)
			if err != nil {
				return nil, defaultErrorHandling(err)
			}
			group = ConvtraqUserGroupTodomainGroup(*g)
			group.IsTraQGroup = true
		} else {
			return nil, defaultErrorHandling(err)
		}
	} else {
		group = db.ConvGroupTodomainGroup(*g)
	}
	return &group, nil
}

func (repo *Repository) GetAllGroups(info *domain.ConInfo) ([]*domain.Group, error) {
	groups := make([]*domain.Group, 0)
	gg, err := repo.GormRepo.GetAllGroups()
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	groups = append(groups, db.ConvSPGroupToSPdomainGroup(gg)...)
	tg, err := repo.TraQRepo.GetAllGroups()
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	dg := ConvSPtraqUserGroupToSPdomainGroup(tg)
	// add IsTraQGroup
	for i := range dg {
		dg[i].IsTraQGroup = true
	}
	// add trap
	groups = append(append(groups, repo.getTraPGroup(info)), dg...)

	return groups, nil
}

func (repo *Repository) GetUserBelongingGroupIDs(userID uuid.UUID, info *domain.ConInfo) ([]uuid.UUID, error) {
	t, err := repo.GormRepo.GetToken(info.ReqUserID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	ggIDs, err := repo.GormRepo.GetBelongGroupIDs(userID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	tgIDs, err := repo.TraQRepo.GetUserBelongingGroupIDs(t, userID)
	if err != nil {
		return nil, defaultErrorHandling(err)
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
		if info.ReqUserID == admin.UserID {
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
	groups, err := repo.TraQRepo.GetAllGroups()
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	names := make([]string, 0)
	for _, g := range groups {
		if g.Type == "grade" && len(g.Members) != 0 {
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
