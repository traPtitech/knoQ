package db

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filter"
	"gorm.io/gorm"
)

func eventFullPreload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Group").Preload("Room").Preload("CreatedBy").
		Preload("Admins").Preload("Admins.UserMeta").Preload("Tags").Preload("Tags.Tag")
}

type WriteEventParams struct {
	domain.WriteEventParams
	CreatedBy uuid.UUID
}

func (repo *GormRepository) CreateEvent(params WriteEventParams) (*Event, error) {
	return createEvent(repo.db, params)
}

func (repo *GormRepository) GetAllEvents(expr filter.Expr) ([]*Event, error) {
	var createFilter func(filter.Expr) (string, []interface{}, error)
	createFilter = func(expr filter.Expr) (string, []interface{}, error) {
		var filterFormat string
		var filterArgs []interface{}

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
				ids, err := getUserBelongingGroupIDs(repo.db, id)
				if err != nil {
					return "", nil, err
				}
				filterFormat = fmt.Sprintf("group_id %v (?)", rel)
				filterArgs = []interface{}{ids}

			default:
				id := e.Value.(uuid.UUID)
				id, ok := e.Value.(uuid.UUID)
				if !ok {
					return "", nil, ErrExpression
				}
				column := map[filter.Attr]string{
					filter.Group: "group_id",
					filter.Tag:   "event_tags.tag_id",
					filter.Event: "id",
				}[e.Attr]
				rel := map[filter.Relation]string{
					filter.Eq:  "=",
					filter.Neq: "!=",
				}[e.Relation]
				filterFormat = fmt.Sprintf("%v %v ?", column, rel)
				filterArgs = []interface{}{id}
			}

		case *filter.LogicOpExpr:
			op := map[filter.LogicOp]string{
				filter.And: "AND",
				filter.Or:  "OR",
			}[e.LogicOp]
			lFilter, lFilterArgs, err := createFilter(e.Lhs)
			if err != nil {
				return "", nil, err
			}
			rFilter, rFilterArgs, err := createFilter(e.Rhs)
			if err != nil {
				return "", nil, err
			}
			filterFormat = fmt.Sprintf("( %v ) %v ( %v )", lFilter, op, rFilter)
			filterArgs = append(lFilterArgs, rFilterArgs...)

		default:
			return "", nil, ErrExpression
		}

		return filterFormat, filterArgs, nil
	}

	filterFormat, filterArgs, err := createFilter(expr)
	if err != nil {
		return nil, err
	}
	return getAllEvents(repo.db, filterFormat, filterArgs)
}

func createEvent(db *gorm.DB, params WriteEventParams) (*Event, error) {
	event := ConvertwriteEventParamsToEvent(params)

	err := db.Create(&event).Error
	return &event, err
}

func updateEvent(db *gorm.DB, eventID uuid.UUID, params WriteEventParams) (*Event, error) {
	event := ConvertwriteEventParamsToEvent(params)
	event.ID = eventID

	err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&event).Error
	return &event, err
}

func getEvent(db *gorm.DB, eventID uuid.UUID) (*Event, error) {
	event := Event{
		ID: eventID,
	}
	cmd := eventFullPreload(db)
	err := cmd.Take(&event).Error
	return &event, err
}

func getAllEvents(db *gorm.DB, query string, args []interface{}) ([]*Event, error) {
	events := make([]*Event, 0)
	cmd := eventFullPreload(db)
	err := cmd.Joins("LEFT JOIN event_tags ON id = event_tags.event_id").
		Where(query, args...).Group("id").Find(&events).Error
	return events, err
}

func addEventTag(db *gorm.DB, eventID uuid.UUID, params domain.EventTagParams) error {
	eventTag := ConvertdomainEventTagParamsToEventTag(params)
	eventTag.EventID = eventID
	return db.Create(&eventTag).Error
}
