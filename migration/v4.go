package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type newEventAdmins struct {
	EventID uuid.UUID `gorm:"type:char(36); primary_key;not null"`
	UserID  uuid.UUID `gorm:"type:char(36); primary_key;not null"`
}

func (*newEventAdmins) TableName() string {
	return "event_admins"
}

type currentEvent struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key"`
	Name        string    `gorm:"type:varchar(32);not null"`
	Description string    `gorm:"type:TEXT"`
	JoinFreely  bool
	CreatedBy   uuid.UUID `gorm:"type:char(36);"`
}

func (*currentEvent) TableName() string {
	return "events"
}

func v4() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "4",
		Migrate: func(db *gorm.DB) error {
			err := db.Migrator().CreateTable(&newEventAdmins{})
			if err != nil {
				return err
			}
			// 作成者を管理ユーザーにする
			events := make([]*currentEvent, 0)
			err = db.Find(&events).Error
			if err != nil {
				return err
			}
			for _, event := range events {
				err = db.Create(&newEventAdmins{
					EventID: event.ID,
					UserID:  event.CreatedBy,
				}).Error
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
}
