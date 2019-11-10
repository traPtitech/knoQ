package repository

import (
	"errors"
	"net/url"
	"strconv"
)

func findRvs(values url.Values) ([]Reservation, error) {
	reservations := []Reservation{}
	cmd := db.Preload("Group").Preload("Group.Members").Preload("Group.CreatedBy").Preload("Room").Preload("CreatedBy")

	if values.Get("id") != "" {
		id, _ := strconv.Atoi(values.Get("id"))
		cmd = cmd.Where("id = ?", id)
	}

	if values.Get("name") != "" {
		cmd = cmd.Where("name LIKE ?", "%"+values.Get("name")+"%")
	}

	if values.Get("traQID") != "" {
		groupsID, err := getGroupIDsBytraQID(values.Get("traQID"))
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
		cmd = cmd.Where("date >= ?", values.Get("date_begin"))
	}
	if values.Get("date_end") != "" {
		cmd = cmd.Where("date <= ?", values.Get("date_end"))
	}

	if err := cmd.Order("date asc").Find(&reservations).Error; err != nil {
		return nil, err
	}

	return reservations, nil
}

// AddCreatedBy add CreatedBy
func (reservation *Reservation) AddCreatedBy() error {
	if err := db.Where("traq_id = ?", reservation.CreatedByRefer).First(&reservation.CreatedBy).Error; err != nil {
		return err
	}
	return nil
}

// 時間が部屋の範囲内か、endがstartの後かどうか確認する
func (rv *Reservation) timeConsistency() error {
	timeStart, err := strToTime(rv.TimeStart)
	if err != nil {
		return err
	}
	timeEnd, err := strToTime(rv.TimeEnd)
	if err != nil {
		return err
	}
	if !(rv.Room.inTime(timeStart) && rv.Room.inTime(timeEnd)) {
		return errors.New("invalid time")
	}
	if !timeStart.Before(timeEnd) {
		return errors.New("invalid time")
	}
	return nil
}
