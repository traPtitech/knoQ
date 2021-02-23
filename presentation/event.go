package presentation

import (
	"fmt"
	"time"

	"github.com/traPtitech/knoQ/domain"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/ical"
)

// EventReqWrite is
//go:generate go run github.com/fuji8/gotypeconverter/cmd/type-converter -s EventReqWrite -d domain.WriteEventParams -o converter.go .
type EventReqWrite struct {
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	AllowTogether bool      `json:"sharedRoom"`
	TimeStart     time.Time `json:"timeStart"`
	TimeEnd       time.Time `json:"timeEnd"`
	RoomID        uuid.UUID `json:"roomId" cvt:"Room"`
	GroupID       uuid.UUID `json:"groupId" cvt:"Group"`
	Tags          []struct {
		Name   string `json:"name"`
		Locked bool   `json:"locked"`
	} `json:"tags"`
}

// EventResOne is experimental
//go:generate go run github.com/fuji8/gotypeconverter/cmd/type-converter -s domain.Event -d EventResOne -o converter.go .
type EventResOne struct {
	domain.Event
}

// EventResMulti is for multiple response
//go:generate go run github.com/fuji8/gotypeconverter/cmd/type-converter -s domain.Event -d EventResMulti -o converter.go .
type EventResMulti struct {
	ID            uuid.UUID         `json:"eventId"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	AllowTogether bool              `json:"sharedRoom"`
	TimeStart     time.Time         `json:"timeStart"`
	TimeEnd       time.Time         `json:"timeEnd"`
	RoomID        uuid.UUID         `json:"roomId" cvt:"Room"`
	GroupID       uuid.UUID         `json:"groupId" cvt:"Group"`
	Place         string            `json:"place"`
	GroupName     string            `json:"groupName"`
	Tags          []domain.EventTag `json:"tags"`
	CreatedBy     uuid.UUID         `json:"createdBy"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
	DeletedAt     *time.Time        `json:"deletedAt,omitempty"`
}

func iCalVeventFormat(e *domain.Event, host string) *ical.Event {
	timeLayout := "20060102T150405Z"
	vevent := ical.NewEvent()
	vevent.AddProperty("uid", e.ID.String())
	vevent.AddProperty("dtstamp", time.Now().UTC().Format(timeLayout))
	vevent.AddProperty("dtstart", e.TimeStart.UTC().Format(timeLayout))
	vevent.AddProperty("dtend", e.TimeEnd.UTC().Format(timeLayout))
	vevent.AddProperty("created", e.CreatedAt.UTC().Format(timeLayout))
	vevent.AddProperty("last-modified", e.UpdatedAt.UTC().Format(timeLayout))
	vevent.AddProperty("summary", e.Name)
	e.Description += "\n\n"
	e.Description += "-----------------------------------\n"
	e.Description += "イベント詳細ページ\n"
	e.Description += fmt.Sprintf("%s/events/%v", host, e.ID)
	vevent.AddProperty("description", e.Description)
	vevent.AddProperty("location", e.Room.Place)
	vevent.AddProperty("organizer", e.CreatedBy.DisplayName)

	return vevent
}

func iCalFormat(events []*domain.Event, host string) *ical.Calendar {
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
		vevent := iCalVeventFormat(e, host)
		c.AddEntry(vevent)
	}
	return c
}
