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
	return tx.Preload("Group").Preload("Group.Members").Preload("Group.Admins").Preload("Group.CreatedBy").
		Preload("Room").Preload("Room.Events").Preload("Room.Admins").Preload("Room.CreatedBy").
		Preload("Admins").Preload("Admins.User").
		Preload("Tags").Preload("Tags.Tag").
		Preload("CreatedBy")
}

type WriteEventParams struct {
	domain.WriteEventParams
	CreatedBy uuid.UUID
}

func (repo *GormRepository) CreateEvent(params WriteEventParams) (*Event, error) {
	e, err := createEvent(repo.db, params)
	return e, defaultErrorHandling(err)
}

func (repo *GormRepository) UpdateEvent(eventID uuid.UUID, params WriteEventParams) (*Event, error) {
	e, err := updateEvent(repo.db, eventID, params)
	return e, defaultErrorHandling(err)
}

func (repo *GormRepository) AddEventTag(eventID uuid.UUID, params domain.EventTagParams) error {
	err := addEventTag(repo.db, eventID, params)
	return defaultErrorHandling(err)
}

func (repo *GormRepository) DeleteEvent(eventID uuid.UUID) error {
	err := deleteEvent(repo.db, eventID)
	return defaultErrorHandling(err)
}

func (repo *GormRepository) DeleteEventTag(eventID uuid.UUID, tagName string, deleteLocked bool) error {
	err := deleteEventTag(repo.db, eventID, tagName, deleteLocked)
	return defaultErrorHandling(err)
}

func (repo *GormRepository) GetEvent(eventID uuid.UUID) (*Event, error) {
	es, err := getEvent(eventFullPreload(repo.db), eventID)
	return es, defaultErrorHandling(err)
}

func (repo *GormRepository) GetAllEvents(expr filter.Expr) ([]*Event, error) {
	filterFormat, filterArgs, err := createFilter(repo.db, expr)
	if err != nil {
		return nil, err
	}
	cmd := eventFullPreload(repo.db)
	es, err := getAllEvents(cmd.Joins(
		"LEFT JOIN event_tags ON id = event_tags.event_id "+
			"LEFT JOIN group_members ON group_id = group_members.group_id "+
			"LEFT JOIN event_admins ON id = event_admins.event_id "),
		filterFormat, filterArgs)
	return es, defaultErrorHandling(err)
}

func createEvent(db *gorm.DB, params WriteEventParams) (*Event, error) {
	event := ConvWriteEventParamsToEvent(params)

	err := db.Create(&event).Error
	return &event, err
}

func updateEvent(db *gorm.DB, eventID uuid.UUID, params WriteEventParams) (*Event, error) {
	event := ConvWriteEventParamsToEvent(params)
	event.ID = eventID

	err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&event).Error
	return &event, err
}

func addEventTag(db *gorm.DB, eventID uuid.UUID, params domain.EventTagParams) error {
	eventTag := ConvdomainEventTagParamsToEventTag(params)
	eventTag.EventID = eventID
	return db.Create(&eventTag).Error
}

func deleteEvent(db *gorm.DB, eventID uuid.UUID) error {
	return db.Delete(&Event{ID: eventID}).Error
}

func deleteEventTag(db *gorm.DB, eventID uuid.UUID, tagName string, deleteLocked bool) error {
	if eventID == uuid.Nil {
		return NewValueError(gorm.ErrRecordNotFound, "eventID")
	}
	eventTag := EventTag{
		EventID: eventID,
		Tag: Tag{
			Name: tagName,
		},
	}
	if !deleteLocked {
		db = db.Where("locked = ?", false)
	}

	return db.Delete(&eventTag).Error
}

func getEvent(db *gorm.DB, eventID uuid.UUID) (*Event, error) {
	event := Event{}
	err := db.Take(&event, eventID).Error
	return &event, err
}

func getAllEvents(db *gorm.DB, query string, args []interface{}) ([]*Event, error) {
	events := make([]*Event, 0)
	err := db.Where(query, args...).Group("id").Order("time_start").Find(&events).Error
	return events, err
}

func createFilter(db *gorm.DB, expr filter.Expr) (string, []interface{}, error) {
	if expr == nil {
		return "", []interface{}{}, nil
	}

	attrMap := map[filter.Attr]string{
		filter.AttrUser:   "group_members.user_id",
		filter.AttrBelong: "group_members.user_id",
		filter.AttrAdmin:  "event_admins.user_id",

		filter.AttrName:      "name",
		filter.AttrGroup:     "group_id",
		filter.AttrRoom:      "room_id",
		filter.AttrTag:       "event_tags.tag_id",
		filter.AttrEvent:     "id",
		filter.AttrTimeStart: "time_start",
		filter.AttrTimeEnd:   "time_end",
	}
	defaultRelationMap := map[filter.Relation]string{
		filter.Eq:       "=",
		filter.Neq:      "!=",
		filter.Greter:   ">",
		filter.GreterEq: ">=",
		filter.Less:     "<",
		filter.LessEq:   "<=",
	}

	var cf func(tx *gorm.DB, e filter.Expr) (string, []interface{}, error)
	cf = func(tx *gorm.DB, e filter.Expr) (string, []interface{}, error) {
		var filterFormat string
		var filterArgs []interface{}

		switch e := e.(type) {
		case *filter.CmpExpr:
			switch e.Attr {
			case filter.AttrName:
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
			case filter.AttrTimeStart:
				fallthrough
			case filter.AttrTimeEnd:
				t, ok := e.Value.(time.Time)
				if !ok {
					return "", nil, ErrExpression
				}
				filterFormat = fmt.Sprintf("%v %v ?", attrMap[e.Attr], defaultRelationMap[e.Relation])
				filterArgs = []interface{}{t}
			default:
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
			lFilter, lFilterArgs, lerr := cf(db, e.Lhs)
			rFilter, rFilterArgs, rerr := cf(db, e.Rhs)

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
	return cf(db, expr)
}
