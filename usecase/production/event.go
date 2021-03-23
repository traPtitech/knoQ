package production

import (
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
)

func (repo *Repository) CreateEvent(params domain.WriteEventParams, info *domain.ConInfo) (*domain.Event, error) {
	// groupの確認
	t, err := repo.gormRepo.GetToken(info.ReqUserID)
	if err != nil {
		return nil, err
	}
	_, err = repo.traQRepo.GetGroup(t, params.GroupID)
	if err != nil {
		return nil, err
	}

	p := db.WriteEventParams{
		WriteEventParams: params,
		CreatedBy:        info.ReqUserID,
	}
	event, err := repo.gormRepo.CreateEvent(p)
	if err != nil {
		return nil, err
	}
	e := db.ConvertEventTodomainEvent(*event)
	return &e, nil
}
