package migration

import (
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

type newGroupAdmins struct {
	GroupID uuid.UUID `gorm:"type:char(36); primary_key;not null"`
	UserID  uuid.UUID `gorm:"type:char(36); primary_key;not null"`
}

func (*newGroupAdmins) TableName() string {
	return "group_admins"
}

type currentGroup struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key"`
	Name        string    `gorm:"type:varchar(32);not null"`
	Description string    `gorm:"type:TEXT"`
	JoinFreely  bool
	CreatedBy   uuid.UUID `gorm:"type:char(36);"`
}

func (*currentGroup) TableName() string {
	return "groups"
}

func v3() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "3",
		Migrate: func(db *gorm.DB) error {
			err := db.CreateTable(&newGroupAdmins{}).Error
			if err != nil {
				return err
			}
			// 作成者を管理ユーザーにする
			groups := make([]*currentGroup, 0)
			err = db.Find(&groups).Error
			if err != nil {
				return err
			}
			for _, group := range groups {
				err = db.Create(&newGroupAdmins{
					GroupID: group.ID,
					UserID:  group.CreatedBy,
				}).Error
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
}
