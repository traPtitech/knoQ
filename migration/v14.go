package migration

import (
	"errors"
	"fmt"
	"time"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// roomsにplace,time_start,time_endの複合ユニークインデックスを追加

type v14Room struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey"`
	Place     string    `gorm:"type:varchar(32); uniqueIndex:idx_place_time_start_time_end"`
	TimeStart time.Time `gorm:"type:DATETIME; index; uniqueIndex:idx_place_time_start_time_end"`
	TimeEnd   time.Time `gorm:"type:DATETIME; index; uniqueIndex:idx_place_time_start_time_end"`
}

type v14UpdateRoom struct {
	ID1       uuid.UUID
	Place1    string
	ID2       uuid.UUID
	Place2    string
	TimeStart time.Time
	TimeEnd   time.Time
}

func (*v14Room) TableName() string {
	return "rooms"
}

func v14() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "14",
		Migrate: func(tx *gorm.DB) error {
			query := `
SELECT
	r1.id AS id1,
	r1.place AS place1,
	r2.id AS id2,
	r2.place AS place2,
	r1.time_start,
	r1.time_end
FROM
	rooms r1
INNER JOIN rooms r2
	ON r1.id < r2.id
	AND r1.place = r2.place
	AND r1.time_start = r2.time_start
	AND r1.time_end = r2.time_end
ORDER BY
	r1.id ASC,
	r2.id ASC
`

			if err := tx.Transaction(func(tx *gorm.DB) error {
				duplicatedRooms := []v14UpdateRoom{}
				if err := tx.Raw(query).Scan(&duplicatedRooms).Error; err != nil {
					return err
				}

				for _, r := range duplicatedRooms {
					i := 1
					for {
						if err := tx.
							Table("rooms").
							Where(
								"place = ? AND time_start = ? AND time_end = ?",
								fmt.Sprintf("%s (%d)", r.Place2, i),
								r.TimeStart,
								r.TimeEnd,
							).
							Take(&map[string]any{}).
							Error; err != nil {
							if errors.Is(err, gorm.ErrRecordNotFound) {
								break
							}
							return err
						}
						i++
					}

					if err := tx.
						Table("rooms").
						Where("id = ?", r.ID2).
						Update("place", fmt.Sprintf("%s (%d)", r.Place2, i)).
						Error; err != nil {
						return err
					}
				}

				if err := tx.AutoMigrate(&v14Room{}); err != nil {
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
