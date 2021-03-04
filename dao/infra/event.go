package infra

import (
	"github.com/traPtitech/knoQ/domain"
)

func (repo *GormRepository) CreateEvent(eventParams domain.WriteEventParams,
	info *domain.ConInfo) (*Event, error) {
	event := ConvertdomainWriteEventParamsToEvent(eventParams)
	err := repo.db.Create(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}
