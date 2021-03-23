package db

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

type WriteEventParams struct {
	domain.WriteEventParams
	CreatedBy uuid.UUID
}

func (repo *GormRepository) CreateEvent(params WriteEventParams) (*Event, error) {
	return createEvent(repo.db, params)
}

func createEvent(db *gorm.DB, params WriteEventParams) (*Event, error) {
	event := ConvertwriteEventParamsToEvent(params)

	err := db.Create(&event).Error
	return &event, err
}

func updateEvent(db *gorm.DB, eventID uuid.UUID, params WriteEventParams) (*Event, error) {
	event := ConvertwriteEventParamsToEvent(params)
	event.ID = eventID

	err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&event).Error
	return &event, err
}

func getEvent(db *gorm.DB, eventID uuid.UUID) (*Event, error) {
	event := Event{
		ID: eventID,
	}
	cmd := db.Preload("Group").Preload("Room").Preload("CreatedBy").
		Preload("Admins").Preload("Admins.UserMeta").Preload("Tags").Preload("Tags.Tag")
	err := cmd.Take(&event).Error
	return &event, err
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
