package repository

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
)

// WriteEventParams is used create and update
type WriteEventParams struct {
	Name          string
	Description   string
	GroupID       uuid.UUID
	RoomID        uuid.UUID
	TimeStart     time.Time
	TimeEnd       time.Time
	CreatedBy     uuid.UUID
	AllowTogether bool
	//Tags          struct {
	//ID     uuid.UUID
	//Locked bool
	//}
}

// WriteTagRelationParams is used create and update
type WriteTagRelationParams struct {
	ID     uuid.UUID
	Locked bool
}

// EventRepository is implemented by GormRepositoy and API repository.
type EventRepository interface {
	CreateEvent(eventParams WriteEventParams) (*Event, error)
	UpdateEvent(eventID uuid.UUID, eventParams WriteEventParams) (*Event, error)
	UpdateTagsInEvent(eventID uuid.UUID, tagsParams []WriteTagRelationParams) error
	AddTagToEvent(eventID uuid.UUID, tagID uuid.UUID, locked bool) error
	//AddEventToFavorites(eventID uuid.UUID, userID uuid.UUID) error
	DeleteEvent(eventID uuid.UUID) error
	// DeleteTagInEvent delete a tag in that Event
	DeleteTagInEvent(eventID uuid.UUID, tagID uuid.UUID, deleteLocked bool) error
	DeleteAllTagInEvent(eventID uuid.UUID) error
	//DeleteEventFavorite(eventID uuid.UUID, userID uuid.UUID) error
	GetEvent(eventID uuid.UUID) (*Event, error)
	GetAllEvents(start *time.Time, end *time.Time) ([]*Event, error)
	GetEventsByGroupIDs(groupIDs []uuid.UUID) ([]*Event, error)
}

// CreateEvent roomが正当かは見る
func (repo *GormRepository) CreateEvent(eventParams WriteEventParams) (*Event, error) {
	event := new(Event)
	err := copier.Copy(&event, eventParams)
	if err != nil {
		return nil, ErrInvalidArg
	}
	// get room
	eventRoom, err := repo.GetRoom(eventParams.RoomID)
	if err != nil {
		return nil, ErrInvalidArg
	}
	event.Room = *eventRoom
	if !event.IsTimeConsistency() {
		return nil, ErrInvalidArg
	}
	err = repo.DB.Create(&event).Error
	return event, err
}

func (repo *GormRepository) UpdateEvent(eventID uuid.UUID, eventParams WriteEventParams) (*Event, error) {
	if eventID == uuid.Nil {
		return nil, ErrNilID
	}

	event := new(Event)
	err := copier.Copy(&event, eventParams)
	if err != nil {
		return nil, ErrInvalidArg
	}
	tx := repo.DB.Begin()
	err = func(tx *gorm.DB) error {
		defer tx.Rollback()
		// delete this event
		result := tx.Delete(&Event{ID: eventID})
		if result.RowsAffected == 0 {
			return ErrNotFound
		}

		// calc time consistency
		event.Room.ID = eventParams.RoomID
		if err := tx.Preload("Events").Take(&event.Room).Error; err != nil {
			return err
		}
		if !event.IsTimeConsistency() {
			return ErrInvalidArg
		}
		return nil
	}(tx)
	if err != nil {
		return nil, err
	}

	// update event
	event.ID = eventID
	err = repo.DB.Save(&event).Error
	return event, err
}

func (repo *GormRepository) UpdateTagsInEvent(eventID uuid.UUID, tagsParams []WriteTagRelationParams) error {
	if eventID == uuid.Nil {
		return ErrNilID
	}
	return repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("event_id = ?", eventID).Delete(&EventTag{}).Error; err != nil {
			return err
		}

		for _, tagParams := range tagsParams {
			eventTag := &EventTag{
				EventID: eventID,
				TagID:   tagParams.ID,
				Locked:  tagParams.Locked,
			}
			if err := tx.Create(&eventTag).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// AddTagToEvent タグを追加。
func (repo *GormRepository) AddTagToEvent(eventID uuid.UUID, tagID uuid.UUID, locked bool) error {
	if eventID == uuid.Nil || tagID == uuid.Nil {
		return ErrNilID
	}
	eventTag := &EventTag{
		EventID: eventID,
		TagID:   tagID,
		Locked:  locked,
	}
	// TODO update event updated_at
	return repo.DB.Create(&eventTag).Error
}

func (repo *GormRepository) DeleteEvent(eventID uuid.UUID) error {
	if eventID == uuid.Nil {
		return ErrNilID
	}
	return repo.DB.Delete(&Event{ID: eventID}).Error
}

func (repo *GormRepository) DeleteTagInEvent(eventID uuid.UUID, tagID uuid.UUID, deleteLocked bool) error {
	if eventID == uuid.Nil || tagID == uuid.Nil {
		return ErrNilID
	}
	cmd := repo.DB
	if !deleteLocked {
		cmd = cmd.Where("locked = ?", false)
	}
	eventTag := &EventTag{
		EventID: eventID,
		TagID:   tagID,
	}
	result := cmd.Delete(&eventTag)
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (repo *GormRepository) DeleteAllTagInEvent(eventID uuid.UUID) error {
	if eventID == uuid.Nil {
		return ErrNilID
	}
	return repo.DB.Where("event_id = ?", eventID).Delete(&EventTag{}).Error
}

func (repo *GormRepository) GetEvent(eventID uuid.UUID) (*Event, error) {
	if eventID == uuid.Nil {
		return nil, ErrNilID
	}

	event := new(Event)
	event.ID = eventID
	cmd := repo.DB.Preload("Room").Preload("Tags")
	err := cmd.First(&event).Error
	return event, err
}

func (repo *GormRepository) GetAllEvents(start *time.Time, end *time.Time) ([]*Event, error) {
	events := make([]*Event, 0)
	cmd := repo.DB.Preload("Room").Preload("Tags")
	if start != nil && !start.IsZero() {
		cmd = cmd.Where("time_end >= ?", start.UTC())
	}
	if end != nil && !end.IsZero() {
		cmd = cmd.Where("time_start <= ?", end.String())
	}
	err := cmd.Debug().Order("time_start").Find(&events).Error
	return events, err

}

func (repo *GormRepository) GetEventsByGroupIDs(groupIDs []uuid.UUID) ([]*Event, error) {
	events := make([]*Event, 0)
	cmd := repo.DB.Preload("Room").Preload("Tags")

	err := cmd.Where("group_id IN (?)", groupIDs).Find(&events).Error
	return events, err
}

func (e *Event) Create() error {
	// groupが存在するかチェックし依存関係を追加する
	//if err := e.Group.Read(); err != nil {
	//return err
	//}
	// roomが存在するかチェックし依存関係を追加する
	//if err := e.Room.Read(); err != nil {
	//return err
	//}

	//err := e.TimeConsistency()
	//if err != nil {
	//return err
	//}

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

	err := tx.Set("gorm:association_save_reference", false).Create(&e).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	// Todo transaction
	//for _, v := range e.Tags {
	//err := e.AddTag(v.ID, v.Locked)
	//if err != nil {
	//tx.Rollback()
	//return err
	//}
	//}

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
	//if err := e.Group.Read(); err != nil {
	//return err
	//}
	// roomが存在するかチェックし依存関係を追加する
	//if err := e.Room.Read(); err != nil {
	//return err
	//}

	//	err := e.TimeConsistency()
	//if err != nil {
	//return err
	//}

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
	//for _, v := range e.Tags {
	//err := e.AddTag(v.ID, v.Locked)
	//if err != nil {
	//tx.Rollback()
	//return err
	//}
	//}

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

// IsTimeConsistency 時間が部屋の範囲内か、endがstartの後か
// available time か確認する
func (e *Event) IsTimeConsistency() bool {
	if !e.Room.InTime(e.TimeStart, e.TimeEnd) {
		return false
	}
	if !e.TimeStart.Before(e.TimeEnd) {
		return false
	}
	return true
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
