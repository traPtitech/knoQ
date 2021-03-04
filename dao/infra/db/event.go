package db

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

type writeEventParams struct {
	domain.WriteEventParams
	CreatedBy uuid.UUID
}

func createEvent(db *gorm.DB, eventParams writeEventParams) (*Event, error) {
	event := new(Event)
	err := db.Create(&event).Error
	if err != nil {
		return nil, err
	}
	return event, nil
}
