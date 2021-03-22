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
	err := cmd.Find(&events).Error
	return events, err
}

func addEventTag(db *gorm.DB, eventID uuid.UUID, tagParams domain.WriteTagRelationParams) error {
	event := Event{ID: eventID}
	tag := ConvertdomainWriteTagRelationParamsToEventTag(tagParams)
	return db.Model(&event).Association("Tags").Append(&tag)
}
