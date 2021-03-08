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

func getAllEvents(db *gorm.DB) ([]*Event, error) {
	events := make([]*Event, 0)
	cmd := db.Preload("Group").Preload("Room").Preload("CreatedBy").
		Preload("Admins").Preload("Admins.UserMeta").Preload("Tags").Preload("Tags.Tag")
	err := cmd.Debug().Find(&events).Error
	return events, err
}
