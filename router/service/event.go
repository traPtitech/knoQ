package service

import (
	"fmt"

	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/parsing"

	repo "github.com/traPtitech/knoQ/repository"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/ical"
)

// GetEventsByUserID get events by userID
func (d Dao) GetEventsByUserID(token string, userID uuid.UUID) ([]*EventRes, error) {
	groupIDs, err := d.GetUserBelongingGroupIDs(token, userID)
	if err != nil {
		return nil, err
	}
	events, err := d.Repo.GetEventsByGroupIDs(groupIDs)
	return FormatEventsRes(events), err
}

// GetiCalByFilter get iCal calendar by specific query
func (d Dao) GetiCalByFilter(token, query, origin string) (*ical.Calendar, error) {
	events, err := d.GetEventsByFilter(token, query)
	if err != nil {
		return nil, err
	}
	c := ical.New()
	ical.NewEvent()
	tz := ical.NewTimezone()
	tz.AddProperty("TZID", "Asia/Tokyo")
	std := ical.NewStandard()
	std.AddProperty("TZOFFSETFROM", "+9000")
	std.AddProperty("TZOFFSETTO", "+9000")
	std.AddProperty("TZNAME", "JST")
	std.AddProperty("DTSTART", "19700101T000000")
	tz.AddEntry(std)
	c.AddEntry(tz)

	for _, e := range events {
		vevent := e.ICal(origin)
		c.AddEntry(vevent)
	}
	return c, nil
}

// GetEventsByFilter get events by specific filter query.
func (d Dao) GetEventsByFilter(token, filterQuery string) ([]*repo.Event, error) {
	expr, err := parsing.Parse(filterQuery)
	if err != nil {
		return nil, fmt.Errorf("%w, %s has '%v'", repo.ErrInvalidArg, filterQuery, err)
	}

	var createFilter func(domain.Expr) (string, []interface{}, error)
	createFilter = func(expr domain.Expr) (string, []interface{}, error) {
		var filter string
		var filterArgs []interface{}

		switch e := expr.(type) {
		case nil:
			filter = ""
			filterArgs = []interface{}{}

		case *domain.CmpExpr:
			id := e.Value.(uuid.UUID)
			switch e.Attr {
			case "user":
				rel := map[domain.Relation]string{
					domain.Eq:  "IN",
					domain.Neq: "NOT IN",
				}[e.Relation]
				ids, err := d.GetUserBelongingGroupIDs(token, id)
				if err != nil {
					return "", nil, err
				}
				filter = fmt.Sprintf("group_id %v (?)", rel)
				filterArgs = []interface{}{ids}

			default:
				column := map[string]string{
					"group": "group_id",
					"tag":   "event_tags.tag_id",
					"event": "id",
				}[e.Attr]
				rel := map[domain.Relation]string{
					domain.Eq:  "=",
					domain.Neq: "!=",
				}[e.Relation]
				filter = fmt.Sprintf("%v %v ?", column, rel)
				filterArgs = []interface{}{id}
			}

		case *domain.LogicOpExpr:
			op := map[domain.LogicOp]string{
				domain.And: "AND",
				domain.Or:  "OR",
			}[e.LogicOp]
			lFilter, lFilterArgs, err := createFilter(e.Lhs)
			if err != nil {
				return "", nil, err
			}
			rFilter, rFilterArgs, err := createFilter(e.Rhs)
			if err != nil {
				return "", nil, err
			}
			filter = fmt.Sprintf("( %v ) %v ( %v )", lFilter, op, rFilter)
			filterArgs = append(lFilterArgs, rFilterArgs...)

		default:
			return "", nil, fmt.Errorf("Unknown expression type")
		}

		return filter, filterArgs, nil
	}

	filter, filterArgs, err := createFilter(expr)
	if err != nil {
		return nil, err
	}
	events, err := d.Repo.GetEventsByFilter(filter, filterArgs)
	if err != nil {
		return nil, err
	}
	return events, err
}
