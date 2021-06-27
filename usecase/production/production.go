package production

import (
	"github.com/traPtitech/knoQ/infra/db"
	"github.com/traPtitech/knoQ/infra/redis"
	"github.com/traPtitech/knoQ/infra/traq"
)

type Repository struct {
	GormRepo  db.GormRepository
	TraQRepo  traq.TraQRepository
	RedisRepo redis.RedisRepository
}

// implements domain
