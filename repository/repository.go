package repository

import (
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
	"github.com/traPtitech/knoQ/infra/traq"
)


func NewRepository(gormRepo db.GormRepository, traQRepo traq.TraQRepository) domain.Repository {
	return &repository{
		GormRepo: gormRepo,
		TraQRepo: traQRepo,
	}
}

type repository struct {
	GormRepo db.GormRepository
	TraQRepo traq.TraQRepository
}

// implements domain
