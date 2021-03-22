package db

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// BeforeCreate is hook
func (e *Event) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}

// BeforeCreate is hook
func (r *Room) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}

// BeforeCreate is hook
func (g *Group) BeforeCreate(tx *gorm.DB) (err error) {
	g.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}

// BeforeCreate is hook
func (t *Tag) BeforeCreate(tx *gorm.DB) (err error) {
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
