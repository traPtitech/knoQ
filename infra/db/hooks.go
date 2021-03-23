package db

import (
	"errors"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// BeforeSave is hook
func (e *Event) BeforeSave(tx *gorm.DB) (err error) {
	if e.ID == uuid.Nil {
		e.ID, err = uuid.NewV4()
		if err != nil {
			return err
		}
	}

	// 時間整合性
	Devent := ConvertEventTodomainEvent(*e)
	if !Devent.TimeConsistency() {
		return NewValueError(ErrTimeConsistency, "timeStart", "timeEnd")
	}
	return nil
}

// BeforeCreate is hook
func (e *Event) BeforeCreate(tx *gorm.DB) (err error) {
	// 時間整合性
	r, err := getRoom(tx.Preload("Events"), e.RoomID)
	if err != nil {
		return err
	}
	e.Room = *r
	Devent := ConvertEventTodomainEvent(*e)
	if !Devent.RoomTimeConsistency() {
		return NewValueError(ErrTimeConsistency, "timeStart", "timeEnd", "room")
	}
	return nil
}

// BeforeUpdate is hook
func (e *Event) BeforeUpdate(tx *gorm.DB) (err error) {
	// 時間整合性
	r, err := getRoom(tx.Preload("Events", "id != ?", e.ID), e.RoomID)
	if err != nil {
		return err
	}
	e.Room = *r
	Devent := ConvertEventTodomainEvent(*e)
	if !Devent.RoomTimeConsistency() {
		return NewValueError(ErrTimeConsistency, "timeStart", "timeEnd", "room")
	}

	// delete current m2m
	err = tx.Where("event_id = ?", e.ID).Delete(&EventTag{}).Error
	if err != nil {
		return err
	}
	err = tx.Where("event_id = ?", e.ID).Delete(&EventAdmin{}).Error
	if err != nil {
		return err
	}

	return nil
}

// BeforeSave is hook
func (et *EventTag) BeforeSave(tx *gorm.DB) (err error) {
	// タグが存在しなければ、作ってイベントにタグを追加する
	// 存在すれば、作らずにイベントにタグを追加する
	tag := Tag{
		Name: et.Tag.Name,
	}
	err = tx.Where(&tag).Take(&tag).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	et.Tag.ID = tag.ID
	return nil
}

// BeforeCreate is hook
func (r *Room) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID != uuid.Nil {
		return nil
	}
	r.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}

// BeforeCreate is hook
func (g *Group) BeforeCreate(tx *gorm.DB) (err error) {
	if g.ID != uuid.Nil {
		return nil
	}
	g.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}

// BeforeCreate is hook
func (t *Tag) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID != uuid.Nil {
		return nil
	}
	t.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}

// BeforeCreate is hook
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID != uuid.Nil {
		return nil
	}
	u.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}
