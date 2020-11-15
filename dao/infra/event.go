package infra

import (
	"room/domain"

	"github.com/jinzhu/copier"
)

func (repo *GormRepository) CreateEvent(eventParams domain.WriteEventParams,
	info *domain.ConInfo) (*Event, error) {
	event := new(Event)
	err := copier.Copy(&event, eventParams)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
