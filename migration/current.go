package migration

import (
	"gopkg.in/gormigrate.v1"
)

// Migrations is all db migrations
func Migrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		v1(),
		v2(),
	}
}
