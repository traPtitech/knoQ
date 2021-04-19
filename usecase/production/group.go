package production

import (
	"errors"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
	"gorm.io/gorm"
)

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

//go:generate gotypeconverter -s v3.UserGroup -d domain.Group -o converter.go .
//go:generate gotypeconverter -s []*v3.UserGroup -d []*domain.Group -o converter.go .

func (repo *Repository) GetGroup(groupID uuid.UUID, info *domain.ConInfo) (*domain.Group, error) {
	var group domain.Group
	g, err := repo.GormRepo.GetGroup(groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			t, err := repo.GormRepo.GetToken(info.ReqUserID)
			if err != nil {
				return nil, defaultErrorHandling(err)
			}
			g, err := repo.TraQRepo.GetGroup(t, groupID)
			if err != nil {
				return nil, defaultErrorHandling(err)
			}
			group = Convv3UserGroupTodomainGroup(*g)
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
	t, err := repo.GormRepo.GetToken(info.ReqUserID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	gg, err := repo.GormRepo.GetAllGroups()
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	groups = append(groups, db.ConvSPGroupToSPdomainGroup(gg)...)
	tg, err := repo.TraQRepo.GetAllGroups(t)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	dg := ConvSPv3UserGroupToSPdomainGroup(tg)
	// add IsTraQGroup
	for i := range dg {
		dg[i].IsTraQGroup = true
	}
	groups = append(groups, dg...)

	return groups, nil
}

func (repo *Repository) GetUserBelongingGroupIDs(userID uuid.UUID, info *domain.ConInfo) ([]uuid.UUID, error) {
	t, err := repo.GormRepo.GetToken(info.ReqUserID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	ggIDs, err := repo.GormRepo.GetUserBelongingGroupIDs(userID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	tgIDs, err := repo.TraQRepo.GetUserBelongingGroupIDs(t, userID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	return append(ggIDs, tgIDs...), nil
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
