package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
)

// Migrations is all db migrations
func Migrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		v1(),
		v2(),
		v3(),
		v4(),
	}
}
