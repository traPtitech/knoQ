package production

import (
	"github.com/traPtitech/knoQ/infra/db"
	"github.com/traPtitech/knoQ/infra/traq"
)

type Repository struct {
	gormRepo db.GormRepository
	traQRepo traq.TraQRepository
}

// implements domain
