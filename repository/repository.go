package repository

import (
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
	"github.com/traPtitech/knoQ/infra/traq"
)

// repository struct の実装を隠蔽する
// あくまで domain.Repository の実装としてのみ存在させる
type Repository interface {
	domain.Repository
}

func NewRepository(gormRepo db.GormRepository, traQRepo traq.TraQRepository) Repository {
	repo := repository{
		GormRepo: gormRepo,
		TraQRepo: traQRepo,
	}
	return repo
}

type repository struct {
	GormRepo db.GormRepository
	TraQRepo traq.TraQRepository
}

// implements domain
