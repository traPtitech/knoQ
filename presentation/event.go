package presentation

import (
	"fmt"
	"net/http"
	"time"

	"github.com/traPtitech/knoQ/domain"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/ical"
)

// EventReqWrite is
//go:generate gotypeconverter -s EventReqWrite -d domain.WriteEventParams -o converter.go .
type EventReqWrite struct {
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	AllowTogether bool        `json:"sharedRoom"`
	TimeStart     time.Time   `json:"timeStart"`
	TimeEnd       time.Time   `json:"timeEnd"`
	RoomID        uuid.UUID   `json:"roomId"`
	Place         string      `json:"place"`
	GroupID       uuid.UUID   `json:"groupId"`
	Admins        []uuid.UUID `json:"admins"`
	Tags          []struct {
		Name   string `json:"name"`
		Locked bool   `json:"locked"`
	} `json:"tags"`
}

type EventTagReq struct {
	Name string `json:"name"`
}

// EventDetailRes is experimental
//go:generate gotypeconverter -s domain.Event -d EventDetailRes -o converter.go .
type EventDetailRes struct {
	ID            uuid.UUID     `json:"eventId"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Room          RoomRes       `json:"room"`
	Group         GroupRes      `json:"group"`
	Place         string        `json:"place" cvt:"Room"`
	GroupName     string        `json:"groupName" cvt:"Group"`
	TimeStart     time.Time     `json:"timeStart"`
	TimeEnd       time.Time     `json:"timeEnd"`
	CreatedBy     uuid.UUID     `json:"createdBy"`
	Admins        []uuid.UUID   `json:"admins"`
	Tags          []EventTagRes `json:"tags"`
	AllowTogether bool          `json:"sharedRoom"`
	Model
}

type EventTagRes struct {
	ID     uuid.UUID `json:"tagId" cvt:"Tag"`
	Name   string    `json:"name" cvt:"Tag"`
	Locked bool      `json:"locked"`
}

// EventRes is for multiple response
//go:generate gotypeconverter -s domain.Event -d EventRes -o converter.go .
//go:generate gotypeconverter -s []*domain.Event -d []EventRes -o converter.go .
type EventRes struct {
	ID            uuid.UUID     `json:"eventId"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	AllowTogether bool          `json:"sharedRoom"`
	TimeStart     time.Time     `json:"timeStart"`
	TimeEnd       time.Time     `json:"timeEnd"`
	RoomID        uuid.UUID     `json:"roomId" cvt:"Room"`
	GroupID       uuid.UUID     `json:"groupId" cvt:"Group"`
	Place         string        `json:"place" cvt:"Room"`
	GroupName     string        `json:"groupName" cvt:"Group"`
	Admins        []uuid.UUID   `json:"admins"`
	Tags          []EventTagRes `json:"tags"`
	CreatedBy     uuid.UUID     `json:"createdBy"`
	Model
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

func ICalFormat(events []*domain.Event, host string) *ical.Calendar {
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

func SchedulerMessageFormat(t time.Time, rooms []*domain.Room, events []*domain.Event, origin string) string {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	date := t.In(jst).Format("01/02(Mon)")
	combined := map[bool]string{
		true:  "(併用可)",
		false: "",
	}
	roomMessage := ""
	publicRoomN := 0
	eventMessage := "本日開催されるイベントは、\n"

	if len(rooms) == 0 {
		roomMessage = fmt.Sprintf("%sの進捗部屋は、予約を取っていないようです。\n", date)
	} else {
		for _, room := range rooms {
			if room.Verified {
				publicRoomN++
				roomMessage += fmt.Sprintf("- %s %s ~ %s\n", room.Place,
					room.TimeStart.In(jst).Format("15:04"), room.TimeEnd.In(jst).Format("15:04"))
			}
		}
		if publicRoomN == 0 {
			roomMessage = fmt.Sprintf("%sの進捗部屋は、予約を取っていないようです。\n", date)
		}
	}

	if len(events) == 0 {
		eventMessage = "本日開催予定のイベントはありません。\n"

	} else {
		for _, event := range events {
			eventMessage += fmt.Sprintf("- [%s](%s/events/%s) %s ~ %s @%s %s\n", event.Name, origin, event.ID,
				event.TimeStart.In(jst).Format("15:04"), event.TimeEnd.In(jst).Format("15:04"),
				event.Room.Place, combined[event.AllowTogether])
		}

	}
	return roomMessage + eventMessage
}

func WebhookMessageFormat(e EventDetailRes, method, origin string) (content string) {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	timeFormat := "01/02(Mon) 15:04"

	switch method {
	case http.MethodPost:
		content = "## イベントが作成されました" + "\n"
	case http.MethodPut:
		content = "## イベントが更新されました" + "\n"
	}

	content += fmt.Sprintf("### [%s](%s/events/%s)", e.Name, origin, e.ID) + "\n"
	content += fmt.Sprintf("- 主催: [%s](%s/groups/%s)", e.GroupName, origin, e.Group.ID) + "\n"
	content += fmt.Sprintf("- 日時: %s ~ %s", e.TimeStart.In(jst).Format(timeFormat), e.TimeEnd.In(jst).Format(timeFormat)) + "\n"
	content += fmt.Sprintf("- 場所: %s", e.Room.Place) + "\n"
	content += "\n"
	content += e.Description

	return
}
