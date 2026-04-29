package db

import (
	"context"
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

func (repo *gormRepository) CreateRoom(ctx context.Context, args domain.CreateRoomArgs) (*domain.Room, error) {
	room, err := createRoom(getTx(ctx, repo.db), args)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	r := ConvRoomTodomainRoom(*room)
	return &r, nil
}

func (repo *gormRepository) UpdateRoom(ctx context.Context, roomID uuid.UUID, args domain.UpdateRoomArgs) (*domain.Room, error) {
	room, err := updateRoom(getTx(ctx, repo.db), roomID, args)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	r := ConvRoomTodomainRoom(*room)
	return &r, nil
}

func (repo *gormRepository) UpdateRoomVerified(ctx context.Context, roomID uuid.UUID, verified bool) error {
	return updateRoomVerified(getTx(ctx, repo.db), roomID, verified)
}

func (repo *gormRepository) DeleteRoom(ctx context.Context, roomID uuid.UUID) error {
	return deleteRoom(getTx(ctx, repo.db), roomID)
}

func (repo *gormRepository) GetRoom(ctx context.Context, roomID uuid.UUID, excludeEventID uuid.UUID) (*domain.Room, error) {
	var room *Room
	var err error
	tx := getTx(ctx, repo.db)
	if excludeEventID == uuid.Nil {
		room, err = getRoom(roomFullPreload(tx), roomID)
	} else {
		room, err = getRoom(roomExcludeEventPreload(tx, excludeEventID), roomID)
	}
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	r := ConvRoomTodomainRoom(*room)
	return &r, nil
}

func (repo *gormRepository) GetAllRooms(ctx context.Context, start, end time.Time, excludeEventID uuid.UUID) ([]*domain.Room, error) {
	var rooms []*Room
	var err error
	tx := getTx(ctx, repo.db)
	if excludeEventID == uuid.Nil {
		rooms, err = getAllRooms(roomFullPreload(tx), start, end)
	} else {
		rooms, err = getAllRooms(roomExcludeEventPreload(tx, excludeEventID), start, end)
	}
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	r := ConvSPRoomToSPdomainRoom(rooms)
	return r, nil
}

func validateRoom(db *gorm.DB, r *Room) (err error) {
	room, err := getRoom(db.Preload("Admins"), r.ID)
	if err != nil {
		return err
	}
	Droom := ConvRoomTodomainRoom(*room)
	if !Droom.AdminsValidation() {
		return NewValueError(ErrNoAdmins, "admins")
	}
	return nil
}

func createRoom(db *gorm.DB, args domain.CreateRoomArgs) (*Room, error) {
	room := ConvCreateRoomParamsToRoom(args)
	var err error
	// IDを新規発行
	room.ID, err = uuid.NewV4()
	if err != nil {
		return nil, err
	}
	err = db.Create(&room).Error
	if err != nil {
		return nil, err
	}
	for _, admin := range room.Admins {
		err = db.Save(&admin).Error
		if err != nil {
			return nil, err
		}
	}

	err = validateRoom(db, &room)
	if err != nil {
		return nil, err
	}

	// 時間整合性は service で確認済み
	return &room, err
}

func updateRoom(db *gorm.DB, roomID uuid.UUID, args domain.UpdateRoomArgs) (*Room, error) {
	room := ConvUpdateRoomParamsToRoom(args)
	room.ID = roomID
	if room.ID == uuid.Nil {
		return nil, ErrRoomUndefined
	}

	// RoomAdmin を更新
	err := db.Where("room_id = ?", room.ID).Delete(&RoomAdmin{}).Error
	if err != nil {
		return nil, err
	}
	// Room を更新
	// 時間整合性は service で確認済み
	err = db.Omit("verified", "CreatedAt").Save(&room).Error
	if err != nil {
		return nil, err
	}
	for _, admin := range room.Admins {
		err = db.Save(&admin).Error
		if err != nil {
			return nil, err
		}
	}
	err = validateRoom(db, &room)
	if err != nil {
		return nil, err
	}
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
