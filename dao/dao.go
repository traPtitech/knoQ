package dao

import (
	"github.com/traPtitech/knoQ/dao/infra/db"
	"github.com/traPtitech/knoQ/dao/infra/traq"
)

type Dao struct {
	gormRepo db.GormRepository
	traQRepo traq.TraQRepository
}
