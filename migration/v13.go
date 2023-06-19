package migration

import (
	"fmt"
	"time"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// eventsにname,time_start,time_endの複合ユニークインデックスを追加

type v13Event struct {
	ID        uuid.UUID `gorm:"type:char(36); primaryKey"`
	Name      string    `gorm:"type:varchar(32); not null; uniqueIndex:idx_name_time_start_time_end"`
	TimeStart time.Time `gorm:"type:DATETIME; index; uniqueIndex:idx_name_time_start_time_end"`
	TimeEnd   time.Time `gorm:"type:DATETIME; index; uniqueIndex:idx_name_time_start_time_end"`
}

func (*v13Event) TableName() string {
	return "events"
}

type v13UpdateEvent struct {
	ID1       uuid.UUID
	Name1     string
	ID2       uuid.UUID
	Name2     string
	TimeStart time.Time
	TimeEnd   time.Time
}

func v13() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "13",
		Migrate: func(tx *gorm.DB) error {
			query := `
SELECT
	e1.id AS id1,
	e1.name AS name1,
	e2.id AS id2,
	e2.name AS name2,
	e1.time_start,
	e1.time_end
FROM
	events e1
INNER JOIN events e2
	ON e1.id < e2.id
	AND e1.name = e2.name
	AND e1.time_start = e2.time_start
	AND e1.time_end = e2.time_end
ORDER BY
	e1.id ASC,
	e2.id ASC
`

			if err := tx.Transaction(func(tx *gorm.DB) error {
				duplicatedEvents := []v13UpdateEvent{}
				if err := tx.Raw(query).Scan(&duplicatedEvents).Error; err != nil {
					return err
				}

				for _, e := range duplicatedEvents {
					i := 1
					for {
						var count int64
						if err := tx.
							Table("events").
							Where(
								"name = ? AND time_start = ? AND time_end = ?",
								fmt.Sprintf("%s (%d)", e.Name2, i),
								e.TimeStart,
								e.TimeEnd,
							).
							Count(&count).
							Error; err != nil {
							return err
						}

						if count == 0 {
							break
						}

						i++
					}

					if err := tx.
						Table("events").
						Where("id = ?", e.ID2).
						Update("name", fmt.Sprintf("%s (%d)", e.Name2, i)).
						Error; err != nil {
						return err
					}
				}

				if err := tx.AutoMigrate(&v13Event{}); err != nil {
					return err
				}

				return nil
			}); err != nil {
				return err
			}

			return nil
		},
	}
}
