package dao

import "github.com/traPtitech/knoQ/dao/infra/db"

type Dao struct {
	gormRepo db.GormRepository
}
