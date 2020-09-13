package service

import (
	"fmt"
	"room/parsing"
	repo "room/repository"

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
	ts, err := parsing.LexAndCheckSyntax(filterQuery)
	if err != nil {
		return nil, fmt.Errorf("%w, %s has '%v'", repo.ErrInvalidArg, filterQuery, err)
	}

	// syntax checkは既にされている。
	filter := ""
	filterArgs := []interface{}{}
	var preAttr string
	for ts.HasNext() {
		t := ts.Next()
		switch t.Kind {
		case parsing.Attr:
			switch t.Value {
			case "user":
				preAttr = "user"
				filter += "group_id"
			case "group":
				preAttr = "group"
				filter += "group_id"
			case "tag":
				preAttr = "tag"
				filter += "event_tags.tag_id"
			case "event":
				preAttr = "event"
				filter += "id"
			}
		case parsing.Or:
			filter += "OR"
		case parsing.And:
			filter += "AND"
		case parsing.Eq:
			if preAttr != "user" {
				filter += "="
			} else if preAttr == "user" {
				filter += "IN"
			}
		case parsing.Neq:
			if preAttr != "user" {
				filter += "!="
			} else if preAttr == "user" {
				filter += "NOT IN"
			}
		case parsing.LParen, parsing.RParen:
			filter += t.Kind.String()
		case parsing.UUID:
			filter += "?"

			id, err := uuid.FromString(t.Value)
			if err != nil {
				return nil, err
			}
			if preAttr != "user" {
				filterArgs = append(filterArgs, id)
			} else if preAttr == "user" {
				ids, err := d.GetUserBelongingGroupIDs(token, id)
				if err != nil {
					return nil, err
				}
				filterArgs = append(filterArgs, ids)
			}
		}
		filter += " "
	}
	events, err := d.Repo.GetEventsByFilter(filter, filterArgs)
	if err != nil {
		return nil, err
	}
	return events, err
}
