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

	if e.RoomID == uuid.Nil {
		if e.Room.Place != "" {
			e.Room.Verified = false
			e.Room.TimeStart = e.TimeStart
			e.Room.TimeEnd = e.TimeEnd
			e.Room.CreatedByRefer = e.CreatedByRefer
			e.Room.Admins = ConvSEventAdminToSRoomAdmin(e.Admins)
		} else {
			return NewValueError(ErrRoomUndefined, "roomID", "place")
		}
	}

	// 時間整合性
	Devent := ConvEventTodomainEvent(*e)
	if !Devent.TimeConsistency() {
		return NewValueError(ErrTimeConsistency, "timeStart", "timeEnd")
	}
	return nil
}

// BeforeCreate is hook
func (e *Event) BeforeCreate(tx *gorm.DB) (err error) {
	// 時間整合性
	r, err := getRoom(tx.Preload("Events"), e.RoomID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 該当する部屋がない場合、部屋時間整合性は調べる必要がない
		// 後で作られる
		return nil
	}
	if err != nil {
		return err
	}
	e.Room = *r
	Devent := ConvEventTodomainEvent(*e)
	if !Devent.RoomTimeConsistency() {
		return NewValueError(ErrTimeConsistency, "timeStart", "timeEnd", "room")
	}
	return nil
}

// BeforeUpdate is hook
func (e *Event) BeforeUpdate(tx *gorm.DB) (err error) {
	// 時間整合性
	r, err := getRoom(tx.Preload("Events", "id != ?", e.ID), e.RoomID)
	if err == nil {
		e.Room = *r
		Devent := ConvEventTodomainEvent(*e)
		if !Devent.RoomTimeConsistency() {
			return NewValueError(ErrTimeConsistency, "timeStart", "timeEnd", "room")
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 該当する部屋がない場合、部屋時間整合性は調べる必要がない
		// 後で作られる
		return err
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

func (e *Event) AfterSave(tx *gorm.DB) (err error) {
	event, err := getEvent(tx.Preload("Admins"), e.ID)
	if err != nil {
		return err
	}
	Devent := ConvEventTodomainEvent(*event)
	if !Devent.AdminsValidation() {
		return NewValueError(ErrNoAdmins, "admins")
	}
	return nil
}

func (e *Event) BeforeDelete(tx *gorm.DB) (err error) {
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
	if et.EventID == uuid.Nil {
		return NewValueError(gorm.ErrRecordNotFound, "eventID")
	}

	// 名前からIDを探す
	// タグが存在しなければ、作ってイベントにタグを追加する
	//（自動で作ることを想定 FullSaveAssociations: true等）
	// 存在すれば、作らずにイベントにタグを追加する
	if et.Tag.ID == uuid.Nil {
		tag := Tag{
			Name: et.Tag.Name,
		}
		err = tx.Where(&tag).Take(&tag).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // 作られる
		}
		if err != nil {
			return err
		}

		et.Tag.ID = tag.ID
	}

	return nil
}

func (et *EventTag) BeforeDelete(tx *gorm.DB) (err error) {
	// タグのIDが空で名前が提供されている場合は、
	// 名前に応じたタグを削除する
	if et.TagID == uuid.Nil && et.Tag.Name != "" {
		tag := Tag{
			Name: et.Tag.Name,
		}
		err = tx.Where(&tag).Take(&tag).Error
		if err != nil {
			return err
		}

		et.TagID = tag.ID
	}
	return nil
}

// BeforeSave is hook
func (r *Room) BeforeSave(tx *gorm.DB) (err error) {
	if r.ID != uuid.Nil {
		return nil
	}
	r.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}

	// 時間整合性
	Droom := ConvRoomTodomainRoom(*r)
	if !Droom.TimeConsistency() {
		return NewValueError(ErrTimeConsistency, "timeStart", "timeEnd")
	}
	return nil
}

func (r *Room) BeforeUpdate(tx *gorm.DB) (err error) {
	err = tx.Where("room_id", r.ID).Delete(&RoomAdmin{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *Room) AfterSave(tx *gorm.DB) (err error) {
	room, err := getRoom(tx.Preload("Admins"), r.ID)
	if err != nil {
		return err
	}
	Droom := ConvRoomTodomainRoom(*room)
	if !Droom.AdminsValidation() {
		return NewValueError(ErrNoAdmins, "admins")
	}
	return nil
}

// BeforeSave is hook
func (g *Group) BeforeSave(tx *gorm.DB) (err error) {
	if g.ID != uuid.Nil {
		return nil
	}
	g.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}

func (g *Group) BeforeUpdate(tx *gorm.DB) (err error) {
	// delete current m2m
	err = tx.Where("group_id = ?", g.ID).Delete(&GroupMember{}).Error
	if err != nil {
		return err
	}
	err = tx.Where("group_id = ?", g.ID).Delete(&GroupAdmin{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (g *Group) AfterSave(tx *gorm.DB) (err error) {
	group, err := getGroup(tx.Preload("Admins"), g.ID)
	if err != nil {
		return err
	}
	Dgroup := ConvGroupTodomainGroup(*group)
	if !Dgroup.AdminsValidation() {
		return NewValueError(ErrNoAdmins, "admins")
	}
	return nil
}

func (g *Group) BeforeDelete(tx *gorm.DB) (err error) {
	// delete current m2m
	err = tx.Where("group_id = ?", g.ID).Delete(&GroupMember{}).Error
	if err != nil {
		return err
	}
	err = tx.Where("group_id = ?", g.ID).Delete(&GroupAdmin{}).Error
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
