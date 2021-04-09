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
	g, err := repo.gormRepo.CreateGroup(p)
	if err != nil {
		return nil, err
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
	g, err := repo.gormRepo.UpdateGroup(groupID, p)
	if err != nil {
		return nil, err
	}
	group := db.ConvGroupTodomainGroup(*g)
	return &group, nil
}

// AddMeToGroup add me to that group if that group is open.
func (repo *Repository) AddMeToGroup(groupID uuid.UUID, info *domain.ConInfo) error {
	if !repo.IsGroupJoinFreely(groupID) {
		return domain.ErrForbidden
	}
	return repo.gormRepo.AddMemberToGroup(groupID, info.ReqUserID)
}

func (repo *Repository) DeleteGroup(groupID uuid.UUID, info *domain.ConInfo) error {
	if !repo.IsGroupAdmins(groupID, info) {
		return domain.ErrForbidden
	}
	return repo.gormRepo.DeleteGroup(groupID)
}

// DeleteMeGroup delete me in that group if that group is open.
func (repo *Repository) DeleteMeGroup(groupID uuid.UUID, info *domain.ConInfo) error {
	if !repo.IsGroupJoinFreely(groupID) {
		return domain.ErrForbidden
	}
	return repo.gormRepo.DeleteMemberOfGroup(groupID, info.ReqUserID)
}

//go:generate gotypeconverter -s v3.UserGroup -d domain.Group -o converter.go .
//go:generate gotypeconverter -s []*v3.UserGroup -d []*domain.Group -o converter.go .

func (repo *Repository) GetGroup(groupID uuid.UUID, info *domain.ConInfo) (*domain.Group, error) {
	var group domain.Group
	g, err := repo.gormRepo.GetGroup(groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			t, err := repo.gormRepo.GetToken(info.ReqUserID)
			if err != nil {
				return nil, err
			}
			g, err := repo.traQRepo.GetGroup(t, groupID)
			if err != nil {
				return nil, err
			}
			group = Convv3UserGroupTodomainGroup(*g)
		} else {
			return nil, err
		}
	} else {
		group = db.ConvGroupTodomainGroup(*g)
	}
	return &group, nil
}

func (repo *Repository) GetAllGroups(info *domain.ConInfo) ([]*domain.Group, error) {
	groups := make([]*domain.Group, 0)
	t, err := repo.gormRepo.GetToken(info.ReqUserID)
	if err != nil {
		return nil, err
	}
	gg, err := repo.gormRepo.GetAllGroups()
	if err != nil {
		return nil, err
	}
	groups = append(groups, db.ConvSPGroupToSPdomainGroup(gg)...)
	tg, err := repo.traQRepo.GetAllGroups(t)
	if err != nil {
		return nil, err
	}
	groups = append(groups, ConvSPv3UserGroupToSPdomainGroup(tg)...)

	return groups, nil
}

func (repo *Repository) GetUserBelongingGroupIDs(userID uuid.UUID, info *domain.ConInfo) ([]uuid.UUID, error) {
	t, err := repo.gormRepo.GetToken(info.ReqUserID)
	if err != nil {
		return nil, err
	}
	ggIDs, err := repo.gormRepo.GetUserBelongingGroupIDs(userID)
	if err != nil {
		return nil, err
	}
	tgIDs, err := repo.traQRepo.GetUserBelongingGroupIDs(t, userID)
	if err != nil {
		return nil, err
	}
	return append(ggIDs, tgIDs...), nil
}

func (repo *Repository) IsGroupAdmins(groupID uuid.UUID, info *domain.ConInfo) bool {
	group, err := repo.gormRepo.GetGroup(groupID)
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
	group, err := repo.gormRepo.GetGroup(groupID)
	if err != nil {
		return false
	}
	return group.JoinFreely
}
