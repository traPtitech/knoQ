package traP

import (
	"github.com/traPtitech/knoQ/infra/db"
	"github.com/traPtitech/knoQ/infra/traq"
)

type Repository struct {
	GormRepo db.GormRepository
	TraQRepo traq.TraQRepository
}

// implements domain
