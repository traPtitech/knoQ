package production

import (
	"errors"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
	"gorm.io/gorm"
)

func (repo *Repository) CreateGroup(groupParams domain.WriteGroupParams, info *domain.ConInfo) (*domain.Group, error) {
	panic("not implemented") // TODO: Implement
}

func (repo *Repository) UpdateGroup(groupID uuid.UUID, groupParams domain.WriteGroupParams, info *domain.ConInfo) (*domain.Group, error) {
	panic("not implemented") // TODO: Implement
}

// AddMeToGroup add me to that group if that group is open.
func (repo *Repository) AddMeToGroup(groupID uuid.UUID, info *domain.ConInfo) error {
	panic("not implemented") // TODO: Implement
}

func (repo *Repository) DeleteGroup(groupID uuid.UUID, info *domain.ConInfo) error {
	panic("not implemented") // TODO: Implement
}

// DeleteMeGroup delete me in that group if that group is open.
func (repo *Repository) DeleteMeGroup(groupID uuid.UUID, info *domain.ConInfo) error {
	panic("not implemented") // TODO: Implement
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
			group = Convertv3UserGroupTodomainGroup(*g)
		} else {
			return nil, err
		}
	} else {
		group = db.ConvertGroupTodomainGroup(*g)
	}
	return &group, nil
}

func (repo *Repository) GetAllGroups(info *domain.ConInfo) ([]*domain.Group, error) {
	panic("not implemented") // TODO: Implement
}

func (repo *Repository) GetUserBelongingGroupIDs(userID uuid.UUID, info *domain.ConInfo) ([]uuid.UUID, error) {
	panic("not implemented") // TODO: Implement
}
