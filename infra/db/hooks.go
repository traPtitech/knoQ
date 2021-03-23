package db

import (
	"errors"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// BeforeCreate is hook
func (e *Event) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID != uuid.Nil {
		return nil
	}
	e.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}

	// タグが存在しなければ、作ってイベントにタグを追加する
	// 存在すれば、作らずにイベントにタグを追加する
	for i, t := range e.Tags {
		tag := Tag{
			Name: t.Tag.Name,
		}
		err := tx.Where(&tag).Take(&tag).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		e.Tags[i].Tag.ID = tag.ID
	}

	// 時間整合性
	r, err := getRoom(tx, e.RoomID)
	if err != nil {
		return err
	}
	e.Room = *r
	Devent := ConvertEventTodomainEvent(*e)
	if !Devent.TimeConsistency() {
		return &ValueError{err: ErrTimeConsistency, args: []string{"timeStart", "timeEnd"}}
	}

	if !Devent.RoomTimeConsistency() {
		return &ValueError{err: ErrTimeConsistency, args: []string{"timeStart", "timeEnd", "roomID"}}
	}

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
