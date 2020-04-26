package service

import (
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

// GetiCalByUserID get iCal calendar by user
func (d Dao) GetiCalByUserID(userID uuid.UUID) (*ical.Calendar, error) {
	// TODO include traQ group
	groupIDs, err := d.Repo.GetUserBelongingGroupIDs(userID)
	if err != nil {
		return nil, err
	}
	events, err := d.Repo.GetEventsByGroupIDs(groupIDs)
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
		vevent := e.ICal()
		c.AddEntry(vevent)
	}
	return c, nil
}
