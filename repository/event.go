package repository

import (
	"errors"
	"net/url"
	"room/utils"
	"strconv"
)

func (e *Event) Create() error {
	// groupが存在するかチェックし依存関係を追加する
	if err := e.Group.Read(); err != nil {
		return err
	}
	// roomが存在するかチェックし依存関係を追加する
	if err := e.Room.Read(); err != nil {
		return err
	}

	// format
	e.Room.Date = e.Room.Date[:10]

	err := e.TimeConsistency()
	if err != nil {
		return err
	}

	err = MatchTags(e.Tags, "event")
	if err != nil {
		return err
	}

	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		dbErrorLog(err)
		return err
	}

	err = tx.Set("gorm:association_save_reference", false).Create(&e).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, v := range e.Tags {
		if err := tx.Create(&EventTag{EventID: e.ID, TagID: v.ID, Locked: v.Locked}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (e *Event) Update() error {
	nowEvent := new(Event)
	nowEvent.ID = e.ID
	if err := nowEvent.Read(); err != nil {
		return err
	}

	// groupが存在するかチェックし依存関係を追加する
	if err := e.Group.Read(); err != nil {
		return err
	}
	// roomが存在するかチェックし依存関係を追加する
	if err := e.Room.Read(); err != nil {
		return err
	}

	// r.Date = 2018-08-10T00:00:00+09:00
	e.Room.Date = e.Room.Date[:10]
	err := e.TimeConsistency()
	if err != nil {
		return err
	}

	e.CreatedBy = nowEvent.CreatedBy
	e.CreatedAt = nowEvent.CreatedAt

	if err := DB.Debug().Save(&e).Error; err != nil {
		return err
	}
	return nil
}

func (e *Event) Delete() error {
	if e.ID == 0 {
		err := errors.New("ID=0. You want to Delete All ?")
		dbErrorLog(err)
		return err
	}
	if err := e.Read(); err != nil {
		return err
	}
	if err := DB.Debug().Delete(&e).Error; err != nil {
		dbErrorLog(err)
		return err
	}
	return nil
}

func (e *Event) Read() error {
	cmd := DB.Preload("Group").Preload("Group.Members").Preload("Room").Preload("Tags")
	if err := cmd.First(&e).Error; err != nil {
		dbErrorLog(err)
		return err
	}
	return nil
}

func FindEvents(values url.Values) ([]Event, error) {
	events := []Event{}
	cmd := DB.Preload("Group").Preload("Group.Members").Preload("Room").Preload("Tags")

	if values.Get("id") != "" {
		id, _ := strconv.Atoi(values.Get("id"))
		cmd = cmd.Where("id = ?", id)
	}

	if values.Get("name") != "" {
		cmd = cmd.Where("name LIKE ?", "%"+values.Get("name")+"%")
	}

	if values.Get("traQID") != "" {
		groupsID, err := GetGroupIDsBytraQID(values.Get("traQID"))
		if err != nil {
			return nil, err
		}
		cmd = cmd.Where("group_id in (?)", groupsID)
	}

	if values.Get("groupid") != "" {
		groupid, _ := strconv.Atoi(values.Get("groupid"))
		cmd = cmd.Where("group_id = ?", groupid)
	}

	if values.Get("roomid") != "" {
		roomid, _ := strconv.Atoi(values.Get("roomid"))
		cmd = cmd.Where("room_id = ?", roomid)
	}

	if values.Get("date_begin") != "" {
		cmd = cmd.Where("rooms.date >= ?", values.Get("date_begin"))
	}
	if values.Get("date_end") != "" {
		cmd = cmd.Where("rooms.date <= ?", values.Get("date_end"))
	}

	// room の日付を見たい
	if err := cmd.Select("events.*").Joins("JOIN rooms on rooms.id = room_id").Find(&events).Error; err != nil {
		dbErrorLog(err)
		return nil, err
	}

	return events, nil
}

func (e *Event) AfterFind() (err error) {
	e.GroupID = 0
	e.RoomID = 0
	return
}

// TimeConsistency 時間が部屋の範囲内か、endがstartの後か
// available time か確認する
func (e *Event) TimeConsistency() error {
	timeStart, err := utils.StrToTime(e.TimeStart)
	if err != nil {
		return err
	}
	timeEnd, err := utils.StrToTime(e.TimeEnd)
	if err != nil {
		return err
	}
	if !e.Room.InTime(timeStart, timeEnd) {
		return errors.New("invalid time")
	}
	if !timeStart.Before(timeEnd) {
		return errors.New("invalid time")
	}
	return nil
}

// GetCreatedBy get who created it
func (rv *Event) GetCreatedBy() (string, error) {
	if err := DB.First(&rv).Error; err != nil {
		dbErrorLog(err)
		return "", err
	}
	return rv.CreatedBy, nil
}
