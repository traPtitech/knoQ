package repository

import (
	"github.com/traPtitech/knoQ/infra/db"
)

type Repository struct {
	GormRepo db.GormRepository
	// TraQRepo infra.TraqRepository
}

// implements domain
