package db

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

func roomExcludeEventPreload(tx *gorm.DB, excludeEventID uuid.UUID) *gorm.DB {
	return tx.Preload("Events", "ID != ?", excludeEventID).Preload("Admins").Preload("CreatedBy")
}

func roomFullPreload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Events").Preload("Admins").Preload("CreatedBy")
}

func (repo *gormRepository) CreateRoom(args domain.CreateRoomArgs) (*domain.Room, error) {
	room, err := createRoom(repo.db, args)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	r := ConvRoomTodomainRoom(*room)
	return &r, nil
}

func (repo *gormRepository) UpdateRoom(roomID uuid.UUID, args domain.UpdateRoomArgs) (*domain.Room, error) {
	room, err := updateRoom(repo.db, roomID, args)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	r := ConvRoomTodomainRoom(*room)
	return &r, nil
}

func (repo *gormRepository) UpdateRoomVerified(roomID uuid.UUID, verified bool) error {
	return updateRoomVerified(repo.db, roomID, verified)
}

func (repo *gormRepository) DeleteRoom(roomID uuid.UUID) error {
	return deleteRoom(repo.db, roomID)
}

func (repo *gormRepository) GetRoom(roomID uuid.UUID, excludeEventID uuid.UUID) (*domain.Room, error) {
	var room *Room
	var err error
	if excludeEventID == uuid.Nil {
		room, err = getRoom(roomFullPreload(repo.db), roomID)
	} else {
		room, err = getRoom(roomExcludeEventPreload(repo.db, excludeEventID), roomID)
	}
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	r := ConvRoomTodomainRoom(*room)
	return &r, nil
}

func (repo *gormRepository) GetAllRooms(start, end time.Time, excludeEventID uuid.UUID) ([]*domain.Room, error) {
	var rooms []*Room
	var err error
	if excludeEventID == uuid.Nil {
		rooms, err = getAllRooms(roomFullPreload(repo.db), start, end)
	} else {
		rooms, err = getAllRooms(roomExcludeEventPreload(repo.db, excludeEventID), start, end)
	}
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	r := ConvSPRoomToSPdomainRoom(rooms)
	return r, nil
}

func createRoom(db *gorm.DB, args domain.CreateRoomArgs) (*Room, error) {
	room := ConvCreateRoomParamsToRoom(args)
	var err error
	// IDを新規発行
	room.ID ,err = uuid.NewV4()
	if err != nil {
		return nil,err
	}
	// 時間整合性は service で確認済み
	err = db.Create(&room).Error
	return &room, err
}

func updateRoom(db *gorm.DB, roomID uuid.UUID, args domain.UpdateRoomArgs) (*Room, error) {
	room := ConvUpdateRoomParamsToRoom(args)
	room.ID = roomID
	if room.ID == uuid.Nil {
		return nil,ErrRoomUndefined
	}

	// BeforeSave, BeforeUpdate が発火
	err := db.Where("room_id", room.ID).Delete(&RoomAdmin{}).Error
	if err!=nil{
		return nil,err
	}
	// 時間整合性は service で確認済み
	err = db.Omit("verified", "CreatedAt").Save(&room).Error

	// err := db.Session(&gorm.Session{FullSaveAssociations: true}).
		// Omit("verified", "CreatedAt").Save(&room).Error
	return &room, err
}

func updateRoomVerified(db *gorm.DB, roomID uuid.UUID, verified bool) error {
	// hooksは発火しない
	return db.Model(&Room{}).Where("id = ?", roomID).UpdateColumn("verified", verified).Error
}

func deleteRoom(db *gorm.DB, roomID uuid.UUID) error {
	room := Room{
		ID: roomID,
	}
	err := db.Delete(&room).Error
	return err
}

func getRoom(db *gorm.DB, roomID uuid.UUID) (*Room, error) {
	room := Room{}
	err := db.Debug().Take(&room, roomID).Error
	return &room, err
}

func getAllRooms(db *gorm.DB, start, end time.Time) ([]*Room, error) {
	rooms := make([]*Room, 0)
	if !start.IsZero() {
		db = db.Where("time_start >= ?", start)
	}
	if !end.IsZero() {
		db = db.Where("time_end <= ?", end)
	}
	err := db.Debug().Order("time_start").Find(&rooms).Error
	return rooms, err
}
