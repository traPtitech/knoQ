package repository

import (
	"sort"
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
	GetEventsByRoomIDs(roomIDs []uuid.UUID) ([]*Event, error)
	GetEventActivities(day int) ([]*Event, error)
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
	if !event.IsTimeConsistency(eventParams.AllowTogether) {
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
		if !event.IsTimeConsistency(eventParams.AllowTogether) {
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

// GetAllEvents start <= time_start <= end なイベントを取得
func (repo *GormRepository) GetAllEvents(start *time.Time, end *time.Time) ([]*Event, error) {
	events := make([]*Event, 0)
	cmd := repo.DB.Preload("Room").Preload("Tags")
	if start != nil && !start.IsZero() {
		cmd = cmd.Where("time_start >= ?", start)
	}
	if end != nil && !end.IsZero() {
		cmd = cmd.Where("time_start <= ?", end)
	}
	err := cmd.Order("time_start").Find(&events).Error
	return events, err

}

func (repo *GormRepository) GetEventsByGroupIDs(groupIDs []uuid.UUID) ([]*Event, error) {
	events := make([]*Event, 0)
	cmd := repo.DB.Preload("Room").Preload("Tags")

	err := cmd.Where("group_id IN (?)", groupIDs).Find(&events).Error
	return events, err
}

func (repo *GormRepository) GetEventActivities(day int) ([]*Event, error) {
	events := make([]*Event, 0)
	now := time.Now()
	period := now.AddDate(0, 0, -1*day)
	cmd := repo.DB.Preload("Room").Preload("Tags")

	err := cmd.Unscoped().Where("created_at > ? ", period).Or("updated_at > ?", period).
		Or("deleted_at > ?", period).Find(&events).Error
	if err != nil {
		return nil, err
	}
	sort.Slice(events, func(i, j int) bool {
		mostRecentI := timeMax(&events[i].CreatedAt, timeMax(&events[i].UpdatedAt, events[i].DeletedAt))
		mostRecentJ := timeMax(&events[j].CreatedAt, timeMax(&events[j].UpdatedAt, events[j].DeletedAt))
		return mostRecentI.Unix() > mostRecentJ.Unix()

	})
	return events, nil
}

func (repo *GormRepository) GetEventsByRoomIDs(roomIDs []uuid.UUID) ([]*Event, error) {
	events := make([]*Event, 0)
	cmd := repo.DB.Preload("Room").Preload("Tags")

	err := cmd.Where("room_id IN (?)", roomIDs).Find(&events).Error
	return events, err

}

func timeMax(a, b *time.Time) *time.Time {
	if a == nil {
		a = new(time.Time)
	}
	if b == nil {
		b = new(time.Time)
	}
	if a.Unix() > b.Unix() {
		return a
	}
	return b
}

// BeforeCreate is gorm hook
func (e *Event) BeforeCreate() (err error) {
	e.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}

// IsTimeConsistency 時間が部屋の範囲内か、endがstartの後か
// available time か確認する
func (e *Event) IsTimeConsistency(allowTogether bool) bool {
	if !e.Room.InTime(e.TimeStart, e.TimeEnd, allowTogether) {
		return false
	}
	if !e.TimeStart.Before(e.TimeEnd) {
		return false
	}
	return true
}
