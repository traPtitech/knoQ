package migration

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type v14Room struct {
	ID             uuid.UUID `gorm:"type:char(36);primaryKey"`
	Place          string    `gorm:"type:varchar(32);"`
	Verified       bool
	TimeStart      time.Time `gorm:"type:DATETIME; index"`
	TimeEnd        time.Time `gorm:"type:DATETIME; index"`
	CreatedByRefer uuid.UUID `gorm:"type:char(36);"`
}

func (*v14Room) TableName() string {
	return "rooms"
}

type v14RoomUser struct {
	RoomID uuid.UUID `gorm:"type:char(36); primaryKey"`
	UserID uuid.UUID `gorm:"type:char(36); primaryKey"`
}

func (*v14RoomUser) TableName() string {
	return "room_admin_users"
}

type v14RoomAdmin struct {
	UserID uuid.UUID `gorm:"type:char(36); primaryKey"`
	RoomID uuid.UUID `gorm:"type:char(36); primaryKey"`
}

func (*v14RoomAdmin) TableName() string {
	return "room_admins"
}

func v14() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "14",
		Migrate: func(db *gorm.DB) error {
			// Step 1: Create the new many-to-many table
			err := db.Migrator().CreateTable(&v14RoomUser{})
			if err != nil {
				return err
			}

			// Step 2: Migrate data from RoomAdmin to RoomUser
			roomAdmins := []v14RoomAdmin{}
			err = db.Find(&roomAdmins).Error
			if err != nil {
				return err
			}

			roomUsers := make([]v14RoomUser, len(roomAdmins))
			for i, admin := range roomAdmins {
				roomUsers[i] = v14RoomUser{
					RoomID: admin.RoomID,
					UserID: admin.UserID,
				}
			}

			if len(roomUsers) > 0 {
				err = db.Create(&roomUsers).Error
				if err != nil {
					return err
				}
			}

			// Step 3: Drop the RoomAdmin table
			return db.Migrator().DropTable(&v14RoomAdmin{})
		},
	}
}
