package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filters"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func eventFullPreload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Group").Preload("Group.Members").Preload("Group.Admins").Preload("Group.CreatedBy").
		Preload("Room").Preload("Room.Events").Preload("Room.Admins").Preload("Room.CreatedBy").
		Preload("Admins").
		Preload("Tags").Preload("Tags.Tag").
		Preload("Attendees").Preload("Attendees.User").
		Preload("CreatedBy")
}

type WriteEventParams struct {
	domain.WriteEventParams
	CreatedBy uuid.UUID
}

func (repo *GormRepository) CreateEvent(params WriteEventParams) (*domain.Event, error) {
	e, err := createEvent(repo.db, params)
	de := ConvEventTodomainEvent(*e)
	return &de, defaultErrorHandling(err)
}

func (repo *GormRepository) UpdateEvent(eventID uuid.UUID, params WriteEventParams) (*domain.Event, error) {
	e, err := updateEvent(repo.db, eventID, params)
	de := ConvEventTodomainEvent(*e)
	return &de, defaultErrorHandling(err)
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

func (repo *GormRepository) UpsertEventSchedule(eventID, userID uuid.UUID, scheduleStatus domain.ScheduleStatus) error {
	err := upsertEventSchedule(repo.db, eventID, userID, scheduleStatus)
	return defaultErrorHandling(err)
}

// group が traq のだと zero 値になる
func (repo *GormRepository) GetEvent(eventID uuid.UUID) (*domain.Event, error) {
	e, err := getEvent(eventFullPreload(repo.db), eventID)
	de := ConvEventTodomainEvent(*e)
	return &de, defaultErrorHandling(err)
}

func (repo *GormRepository) GetAllEvents(expr filters.Expr) ([]*domain.Event, error) {
	filterFormat, filterArgs, err := createEventFilter(expr)
	if err != nil {
		return nil, err
	}
	cmd := eventFullPreload(repo.db)
	es, err := getEvents(cmd.Joins(
		"LEFT JOIN event_tags ON events.id = event_tags.event_id "+
			"LEFT JOIN group_member ON events.group_id = group_member.group_id "+
			"LEFT JOIN event_admin ON events.id = event_admin.event_id "+
			"LEFT JOIN event_attendees ON events.id = event_attendees.event_id"), filterFormat, filterArgs)
	des := ConvSPEventToSPdomainEvent(es)
	return des, defaultErrorHandling(err)
}

func createEvent(db *gorm.DB, params WriteEventParams) (*Event, error) {
	event := ConvWriteEventParamsToEvent(params)

	err := db.Create(&event).Error
	return &event, err
}

func updateEvent(db *gorm.DB, eventID uuid.UUID, params WriteEventParams) (*Event, error) {
	event := ConvWriteEventParamsToEvent(params)
	event.ID = eventID

	err := db.Session(&gorm.Session{FullSaveAssociations: true}).
		Omit("CreatedAt").Save(&event).Error
	return &event, err
}

func addEventTag(db *gorm.DB, eventID uuid.UUID, params domain.EventTagParams) error {
	eventTag := ConvdomainEventTagParamsToEventTag(params)
	eventTag.EventID = eventID
	err := db.Create(&eventTag).Error
	if errors.Is(defaultErrorHandling(err), ErrDuplicateEntry) {
		return db.Omit("CreatedAt").Save(&eventTag).Error
	}
	return err
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
		Tag: &Tag{
			Name: tagName,
		},
	}
	if !deleteLocked {
		db = db.Where("locked = ?", false)
	}

	return db.Delete(&eventTag).Error
}

func upsertEventSchedule(tx *gorm.DB, eventID, userID uuid.UUID, schedule domain.ScheduleStatus) error {
	if eventID == uuid.Nil {
		return NewValueError(gorm.ErrRecordNotFound, "eventID")
	}
	eventAttendee := EventAttendee{
		UserID:   userID,
		EventID:  eventID,
		Schedule: int(schedule),
	}

	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "event_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"schedule"}),
	}).Create(&eventAttendee).Error
}

func getEvent(db *gorm.DB, eventID uuid.UUID) (*Event, error) {
	event := Event{}
	err := db.Take(&event, eventID).Error
	return &event, err
}

func getEvents(db *gorm.DB, query string, args []interface{}) ([]*Event, error) {
	events := make([]*Event, 0)
	err := db.Where(query, args...).Group("id").Order("time_start").Find(&events).Error
	return events, err
}

func createEventFilter(expr filters.Expr) (string, []interface{}, error) {
	if expr == nil {
		return "", []interface{}{}, nil
	}

	attrMap := map[filters.Attr]string{
		filters.AttrUser:     "event_attendees.user_id",
		filters.AttrBelong:   "group_member.user_id",
		filters.AttrAdmin:    "event_admin.user_id",
		filters.AttrAttendee: "event_attendees.user_id",

		filters.AttrName:      "events.name",
		filters.AttrGroup:     "events.group_id",
		filters.AttrRoom:      "events.room_id",
		filters.AttrTag:       "event_tags.tag_id",
		filters.AttrEvent:     "events.id",
		filters.AttrTimeStart: "events.time_start",
		filters.AttrTimeEnd:   "events.time_end",
	}
	defaultRelationMap := map[filters.Relation]string{
		filters.Eq:       "=",
		filters.Neq:      "!=",
		filters.Greter:   ">",
		filters.GreterEq: ">=",
		filters.Less:     "<",
		filters.LessEq:   "<=",
	}

	var cf func(e filters.Expr) (string, []interface{}, error)
	cf = func(e filters.Expr) (string, []interface{}, error) {
		var filterFormat string
		var filterArgs []interface{}

		switch e := e.(type) {
		case *filters.CmpExpr:
			switch e.Attr {
			case filters.AttrName:
				name, ok := e.Value.(string)
				if !ok {
					return "", nil, ErrExpression
				}
				rel := map[filters.Relation]string{
					filters.Eq:  "=",
					filters.Neq: "!=",
				}[e.Relation]
				filterFormat = fmt.Sprintf("events.name %v ?", rel)
				filterArgs = []interface{}{name}
			case filters.AttrTimeStart:
				fallthrough
			case filters.AttrTimeEnd:
				t, ok := e.Value.(time.Time)
				if !ok {
					return "", nil, ErrExpression
				}
				filterFormat = fmt.Sprintf("%v %v ?", attrMap[e.Attr], defaultRelationMap[e.Relation])
				filterArgs = []interface{}{t}
			case filters.AttrUser:
				fallthrough
			case filters.AttrAttendee:
				id, ok := e.Value.(uuid.UUID)
				if !ok {
					return "", nil, ErrExpression
				}
				filterFormat = fmt.Sprintf("%v %v ? AND event_attendees.schedule != %v", attrMap[e.Attr], defaultRelationMap[e.Relation], domain.Absent)
				filterArgs = []interface{}{id}
			default:
				id, ok := e.Value.(uuid.UUID)
				if !ok {
					return "", nil, ErrExpression
				}
				filterFormat = fmt.Sprintf("%v %v ?", attrMap[e.Attr], defaultRelationMap[e.Relation])
				filterArgs = []interface{}{id}
			}

		case *filters.LogicOpExpr:
			op := map[filters.LogicOp]string{
				filters.And: "AND",
				filters.Or:  "OR",
			}[e.LogicOp]
			lFilter, lFilterArgs, lerr := cf(e.LHS)
			rFilter, rFilterArgs, rerr := cf(e.RHS)

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
			filterArgs = lFilterArgs
			filterArgs = append(filterArgs, rFilterArgs...)

		default:
			return "", nil, ErrExpression
		}

		return filterFormat, filterArgs, nil
	}
	return cf(expr)
}
