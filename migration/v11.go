package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	newModel "github.com/traPtitech/knoQ/migration/v8"
)

// v11 特権ユーザーを verifed = true な room の管理者にする
func v11() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "11",
		Migrate: func(db *gorm.DB) error {
			previlegedUsers := make([]*newModel.User, 0)
			err := db.Where("privilege = ?", true).Find(&previlegedUsers).Error
			if err != nil {
				return err
			}

			publicRooms := make([]*newModel.Room, 0)
			err = db.Where("verifed = ?", true).Find(&publicRooms).Error
			if err != nil {
				return err
			}

			roomAdmins := make([]*newModel.RoomAdmin, 0)
			for _, room := range publicRooms {
				for _, user := range previlegedUsers {
					roomAdmins = append(roomAdmins, &newModel.RoomAdmin{
						RoomID: room.ID,
						UserID: user.ID,
					})
				}
			}

			return db.Create(&roomAdmins).Error
		},
	}
}
