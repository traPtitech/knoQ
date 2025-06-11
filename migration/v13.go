package migration

import (
	"database/sql"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- 移行前の一時的な構造体定義 ---
// データ移行の読み取り元として使用します。
type v13OldRoom struct {
	ID       uuid.UUID `gorm:"type:char(36);primaryKey"`
	Place    string    `gorm:"type:varchar(32);"`
	Verified bool
}

func (v13OldRoom) TableName() string {
	return "rooms"
}

/*
古い Event
```
	RoomID         uuid.UUID `gorm:"type:char(36); not null; index"` // ここを null も許容するようにしたい
	Room           Room      `gorm:"foreignKey:RoomID; constraint:OnDelete:CASCADE;" cvt:"write:Place"`
	TimeStart      time.Time `gorm:"type:DATETIME; index"`
	TimeEnd        time.Time `gorm:"type:DATETIME; index"`
```
*/

type v13OldEvent struct {
	ID     uuid.UUID  `gorm:"type:char(36); primaryKey"`
	RoomID uuid.UUID  `gorm:"type:char(36); not null; index"`
	Room   v13OldRoom `gorm:"foreignKey:RoomID;"`
}

func (v13OldEvent) TableName() string {
	return "events"
}

// --- 移行後の最終的な構造体定義 ---
/*
	Place          string    `gorm:"type:varchar(32);"`
	Verified       bool
	Events         []Event   `gorm:"->; constraint:-"` // readOnly
	Admins         []RoomAdmin
*/
// type v13NewRoom struct {
// 	ID   uuid.UUID `gorm:"type:char(36);primaryKey"`
// 	Name string    `gorm:"type:varchar(32);"` // placeからrename
// }

// func (v13NewRoom) TableName() string {
// 	return "rooms"
// }

type v13NewEvent struct {
	ID          uuid.UUID      `gorm:"type:char(36);primaryKey"`
	IsRoomEvent bool           `gorm:"<-:create;not null"`   // 新規カラム
	Venue       sql.NullString `gorm:"type:varchar(255)"`    // 新規カラム
	RoomID      uuid.NullUUID  `gorm:"type:char(36); index"` // NULL許容に変更
}

func (v13NewEvent) TableName() string {
	return "events"
}

type v13RoomAdmin struct{}

func (a v13RoomAdmin) TableName() string {
	return "room_admins"
}

// v13 は Room を進捗部屋専用の構造体とし、非進捗部屋の情報を Event に移行するマイグレーションです。
func v13() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "13",
		Migrate: func(db *gorm.DB) error {
			// 1a. カラム`is_room_event`の追加
			//     まず、カラムが存在しないことを確認します。
			if !db.Migrator().HasColumn(&v13NewEvent{}, "IsRoomEvent") {
				// カラムが存在しない場合のみ、追加処理を実行します。
				if err := db.Migrator().AddColumn(&v13NewEvent{}, "IsRoomEvent"); err != nil {
					return err
				}
			}

			// 1b. カラム`venue`の追加
			if !db.Migrator().HasColumn(&v13NewEvent{}, "Venue") {
				if err := db.Migrator().AddColumn(&v13NewEvent{}, "Venue"); err != nil {
					return err
				}
			}

			// - `events.room_id` を NULL 許容に変更
			if err := db.Migrator().AlterColumn(&v13NewEvent{}, "RoomID"); err != nil {
				return err
			}

			// 2. データ移行
			var oldEvents []v13OldEvent
			if err := db.Preload("Room").Find(&oldEvents).Error; err != nil {
				return err
			}

			for _, event := range oldEvents {
				if event.Room.Verified {
					// 進捗部屋 (Verified = true) の場合: is_room_eventをtrueに
					if err := db.Model(&v13NewEvent{}).Where("id = ?", event.ID).Update("is_room_event", true).Error; err != nil {
						return err
					}
				} else {
					// 非進捗部屋 (Verified = false) の場合:
					// is_room_eventをfalse, venueに場所名をコピー, room_idをNULLに
					updates := map[string]interface{}{
						"is_room_event": false,
						"venue":         event.Room.Place,
						"room_id":       nil,
					}
					if err := db.Model(&v13NewEvent{}).Where("id = ?", event.ID).Updates(updates).Error; err != nil {
						return err
					}
				}
			}
			// 3. 不要になった非進捗部屋と、それに関連するデータを削除

			// 3a. まず、削除対象となる非進捗部屋のIDリストを取得
			var nonProgressRoomIDs []uuid.UUID
			if err := db.Model(&v13OldRoom{}).Where("verified = ?", false).Pluck("id", &nonProgressRoomIDs).Error; err != nil {
				return err
			}

			if len(nonProgressRoomIDs) > 0 {
				// 3b. 次に、`room_admins` テーブルから関連するレコードを削除 (子テーブルから先に)
				if err := db.Where("room_id IN (?)", nonProgressRoomIDs).Delete(&v13RoomAdmin{}).Error; err != nil {
					return err
				}
			}

			// 3c. 最後に、不要になった非進捗部屋のRoomレコード本体を削除 (親テーブルを後に)
			if err := db.Where("verified = ?", false).Delete(&v13OldRoom{}).Error; err != nil {
				return err
			}

			// 4. 不要になった`rooms.verified`カラムを削除
			if db.Migrator().HasColumn(&v13OldRoom{}, "verified") {
				if err := db.Migrator().DropColumn(&v13OldRoom{}, "verified"); err != nil {
					return err
				}
			}

			// - `rooms`テーブルの`place`カラムの型などを`name`カラムに合わせて更新
			// 5. `rooms.place`カラムを`name`にリネーム
			if db.Migrator().HasColumn(&v13OldRoom{}, "place") {
				if err := db.Migrator().RenameColumn(&v13OldRoom{}, "place", "name"); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
