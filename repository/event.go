package repository

import (
	"errors"
	"net/url"
	"room/utils"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
)

// WriteEventParams is used create and update
type WriteEventParams struct {
}

// EventRepository is implemented by GormRepositoy and API repository.
type EventRepository interface {
	CreateEvent(eventParams WriteEventParams) (*Event, error)
	UpdateEvent(eventID uuid.UUID, eventParams WriteEventParams) (*Event, error)
	AddTagToEvent(eventID uuid.UUID, tagID uuid.UUID) error
	AddEventToFavorites(eventID uuid.UUID, userID uuid.UUID) error
	DeleteEvent(eventID uuid.UUID) error
	// DeleteTagInEvent delete a tag in that Event if that tag is locked == false
	DeleteTagInEvent(eventID uuid.UUID, tagID uuid.UUID) error
	DeleteEventFavorite(eventID uuid.UUID, userID uuid.UUID) error
	GetEvent(eventID uuid.UUID) (*Event, error)
	GetAllEvents(start *time.Time, end *time.Time) ([]*Event, error)
	GetEventsByGroupIDs(groupIDs []uuid.UUID) ([]*Event, error)
}

func (e *Event) Create() error {
	// groupが存在するかチェックし依存関係を追加する
	if err := e.Group.Read(); err != nil {
		return err
	}
	// roomが存在するかチェックし依存関係を追加する
	if err := e.Room.Read(); err != nil {
		return err
	}

	err := e.TimeConsistency()
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
	// Todo transaction
	for _, v := range e.Tags {
		err := e.AddTag(v.ID, v.Locked)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (e *Event) Read() error {
	cmd := DB.Preload("Group").Preload("Group.Members").Preload("Room").Preload("Tags")
	if err := cmd.First(&e).Error; err != nil {
		dbErrorLog(err)
		return err
	}
	return nil
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

	err := e.TimeConsistency()
	if err != nil {
		return err
	}

	e.CreatedAt = nowEvent.CreatedAt
	e.CreatedBy = nowEvent.CreatedBy

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

	if err := tx.Debug().Save(&e).Error; err != nil {
		tx.Rollback()
		return err
	}

	// delete now all tags
	if err := tx.Model(&nowEvent).Association("Tags").Clear().Error; err != nil {
		tx.Rollback()
		return err
	}
	// Todo transaction
	for _, v := range e.Tags {
		err := e.AddTag(v.ID, v.Locked)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (e *Event) Delete() error {
	if uuid.Nil == e.ID {
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

// BeforeCreate is gorm hook
func (e *Event) BeforeCreate() (err error) {
	e.ID, err = uuid.NewV4()
	if err != nil {
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
func (rv *Event) GetCreatedBy() (uuid.UUID, error) {
	if err := DB.First(&rv).Error; err != nil {
		dbErrorLog(err)
		return uuid.Nil, err
	}
	return rv.CreatedBy, nil
}

// AddTag add tag
func (e *Event) AddTag(tagID uuid.UUID, locked bool) error {
	tag := new(Tag)
	tag.ID = tagID
	if err := DB.First(&tag).Error; err != nil {
		dbErrorLog(err)
		return err
	}
	if err := DB.Create(&EventTag{EventID: e.ID, TagID: tag.ID, Locked: locked}).Error; err != nil {
		return err
	}
	return nil
}

// DeleteTag delete unlocked tag.
func (e *Event) DeleteTag(tagID uuid.UUID) error {
	eventTag := new(EventTag)
	eventTag.TagID = tagID
	eventTag.EventID = e.ID
	if err := DB.Debug().First(&eventTag).Error; err != nil {
		return err
	}
	if eventTag.Locked {
		return errors.New("this tag is locked")
	}
	if err := DB.Debug().Where("locked = ?", false).Delete(&EventTag{EventID: e.ID, TagID: eventTag.TagID}).Error; err != nil {
		return err
	}
	return nil
}
