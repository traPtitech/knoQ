package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func v13() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "13",
		Migrate: func(db *gorm.DB) error {
			return db.Exec(
				`
SET @sql = (
	SELECT IF(
		COUNT(*) > 0,
		'ALTER TABLE events DROP CONSTRAINT idx_name_time_start_time_end',
		'SELECT "Constraint idx_name_time_start_time_end does not exist thus not dropping it"'
	) FROM information_schema.TABLE_CONSTRAINTS
	WHERE CONSTRAINT_NAME = 'idx_name_time_start_time_end' AND TABLE_NAME = 'events'
);

PREPARE statement FROM @sql;
EXECUTE statement;
DEALLOCATE PREPARE statement;
`,
			).Error
		},
	}
}
