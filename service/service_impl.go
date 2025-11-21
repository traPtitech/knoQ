package service

import (
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
	"github.com/traPtitech/knoQ/infra/traq"
)

type service struct {
	GormRepo *db.GormRepository
	TraQRepo *traq.TraQRepository
}

// implements domain

func NewService(gormRepo *db.GormRepository, traqRepo *traq.TraQRepository) domain.Service {
	return &service{GormRepo: gormRepo, TraQRepo: traqRepo}
}
