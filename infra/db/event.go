package db

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filter"
	"gorm.io/gorm"
)

func eventFullPreload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Group").Preload("Room").Preload("CreatedBy").
		Preload("Admins").Preload("Admins.User").Preload("Tags").Preload("Tags.Tag")
}

type WriteEventParams struct {
	domain.WriteEventParams
	CreatedBy uuid.UUID
}

func (repo *GormRepository) CreateEvent(params WriteEventParams) (*Event, error) {
	return createEvent(repo.db, params)
}

func (repo *GormRepository) UpdateEvent(eventID uuid.UUID, params WriteEventParams) (*Event, error) {
	return updateEvent(repo.db, eventID, params)
}

func (repo *GormRepository) AddEventTag(eventID uuid.UUID, params domain.EventTagParams) error {
	return addEventTag(repo.db, eventID, params)
}

func (repo *GormRepository) DeleteEvent(eventID uuid.UUID) error {
	return deleteEvent(repo.db, eventID)
}

func (repo *GormRepository) DeleteEventTag(eventID uuid.UUID, tagName string) error {
	return deleteEventTag(repo.db, eventID, tagName)
}

func (repo *GormRepository) GetEvent(eventID uuid.UUID) (*Event, error) {
	return getEvent(eventFullPreload(repo.db), eventID)
}

func (repo *GormRepository) GetAllEvents(expr filter.Expr) ([]*Event, error) {
	filterFormat, filterArgs, err := createFilter(repo.db, expr)
	if err != nil {
		return nil, err
	}
	cmd := eventFullPreload(repo.db)
	return getAllEvents(cmd.Joins("LEFT JOIN event_tags ON id = event_tags.event_id"),
		filterFormat, filterArgs)
}

func createEvent(db *gorm.DB, params WriteEventParams) (*Event, error) {
	event := ConvertWriteEventParamsToEvent(params)

	err := db.Create(&event).Error
	return &event, err
}

func updateEvent(db *gorm.DB, eventID uuid.UUID, params WriteEventParams) (*Event, error) {
	event := ConvertWriteEventParamsToEvent(params)
	event.ID = eventID

	err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&event).Error
	return &event, err
}

func addEventTag(db *gorm.DB, eventID uuid.UUID, params domain.EventTagParams) error {
	eventTag := ConvertdomainEventTagParamsToEventTag(params)
	eventTag.EventID = eventID
	return db.Create(&eventTag).Error
}

func deleteEvent(db *gorm.DB, eventID uuid.UUID) error {
	return db.Where("id = ?", eventID).Delete(&Event{}).Error
}

func deleteEventTag(db *gorm.DB, eventID uuid.UUID, tagName string) error {
	eventTag := EventTag{
		EventID: eventID,
		Tag: Tag{
			Name: tagName,
		},
	}
	return db.Where("locked = ?", false).Delete(&eventTag).Error
}

func getEvent(db *gorm.DB, eventID uuid.UUID) (*Event, error) {
	event := Event{
		ID: eventID,
	}
	err := db.Take(&event).Error
	return &event, err
}

func getAllEvents(db *gorm.DB, query string, args []interface{}) ([]*Event, error) {
	events := make([]*Event, 0)
	err := db.Where(query, args...).Group("id").Order("time_start").Find(&events).Error
	return events, err
}

func createFilter(db *gorm.DB, expr filter.Expr) (string, []interface{}, error) {
	var filterFormat string
	var filterArgs []interface{}

	attrMap := map[filter.Attr]string{
		filter.User:      "group_id",
		filter.Name:      "name",
		filter.Group:     "group_id",
		filter.Room:      "room_id",
		filter.Tag:       "event_tags.tag_id",
		filter.Event:     "id",
		filter.TimeStart: "time_start",
		filter.TimeEnd:   "time_end",
	}
	defaultRelationMap := map[filter.Relation]string{
		filter.Eq:       "=",
		filter.Neq:      "!=",
		filter.Greter:   ">",
		filter.GreterEq: ">=",
		filter.Less:     "<",
		filter.LessEq:   "<=",
	}

	switch e := expr.(type) {
	case nil:
		filterFormat = ""
		filterArgs = []interface{}{}

	case *filter.CmpExpr:
		switch e.Attr {
		case filter.User:
			id, ok := e.Value.(uuid.UUID)
			if !ok {
				return "", nil, ErrExpression
			}
			rel := map[filter.Relation]string{
				filter.Eq:  "IN",
				filter.Neq: "NOT IN",
			}[e.Relation]
			ids, err := getUserBelongingGroupIDs(db, id)
			if err != nil {
				return "", nil, err
			}

			filterFormat = fmt.Sprintf("%s %v (?)", attrMap[e.Attr], rel)
			filterArgs = []interface{}{ids}

		case filter.Name:
			name, ok := e.Value.(string)
			if !ok {
				return "", nil, ErrExpression
			}
			rel := map[filter.Relation]string{
				filter.Eq:  "=",
				filter.Neq: "!=",
			}[e.Relation]
			filterFormat = fmt.Sprintf("name %v ?", rel)
			filterArgs = []interface{}{name}
		case filter.TimeStart:
			fallthrough
		case filter.TimeEnd:
			t, ok := e.Value.(time.Time)
			if !ok {
				return "", nil, ErrExpression
			}
			filterFormat = fmt.Sprintf("%v %v ?", attrMap[e.Attr], defaultRelationMap[e.Relation])
			filterArgs = []interface{}{t}
		default:
			id := e.Value.(uuid.UUID)
			id, ok := e.Value.(uuid.UUID)
			if !ok {
				return "", nil, ErrExpression
			}
			filterFormat = fmt.Sprintf("%v %v ?", attrMap[e.Attr], defaultRelationMap[e.Relation])
			filterArgs = []interface{}{id}
		}

	case *filter.LogicOpExpr:
		op := map[filter.LogicOp]string{
			filter.And: "AND",
			filter.Or:  "OR",
		}[e.LogicOp]
		lFilter, lFilterArgs, lerr := createFilter(db, e.Lhs)
		rFilter, rFilterArgs, rerr := createFilter(db, e.Rhs)

		if lerr != nil && rerr != nil {
			return "", nil, ErrExpression
		}
		if lerr != nil {
			return rFilter, rFilterArgs, nil
		}
		if rerr != nil {
			return lFilter, lFilterArgs, nil
		}

		filterFormat = fmt.Sprintf("( %v ) %v ( %v )", lFilter, op, rFilter)
		filterArgs = append(lFilterArgs, rFilterArgs...)

	default:
		return "", nil, ErrExpression
	}

	return filterFormat, filterArgs, nil
}
