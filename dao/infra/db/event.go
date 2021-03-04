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

// BeforeCreate is hook
func (e *Event) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}

func createEvent(db *gorm.DB, eventParams writeEventParams) (*Event, error) {
	event := ConvertwriteEventParamsToEvent(eventParams)
	err := db.Create(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}
