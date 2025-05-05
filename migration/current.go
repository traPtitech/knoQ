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
		v5(),
		v6(),
		v7(),
		v8(),
		v9(),
		v10(),
		v11(),
		v12(),
		v13(),
		v14(),
		v15(),
		v16(),
	}
}
