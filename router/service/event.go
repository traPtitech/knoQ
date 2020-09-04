package service

import (
	"room/parsing"

	"github.com/gofrs/uuid"
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

// GetEventsByFilter get events by specific filter query.
func (d Dao) GetEventsByFilter(token, filterQuery string) ([]*EventRes, error) {
	ts, err := parsing.LexAndCheckSyntax(filterQuery)
	if err != nil {
		return nil, err
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
	return FormatEventsRes(events), err
}
