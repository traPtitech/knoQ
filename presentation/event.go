package presentation

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/traPtitech/knoQ/domain"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/ical"
)

type ScheduleStatus int

const (
	Pending ScheduleStatus = iota + 1
	Attendance
	Absent
)

// EventReqWrite is
//go:generate gotypeconverter -s EventReqWrite -d domain.WriteEventParams -o converter.go .
type EventReqWrite struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	TimeStart   time.Time   `json:"timeStart"`
	TimeEnd     time.Time   `json:"timeEnd"`
	Rooms       []EventRoom `json:"rooms"`
	GroupID     uuid.UUID   `json:"groupId"`
	Admins      []uuid.UUID `json:"admins"`
	Tags        []struct {
		Name   string `json:"name"`
		Locked bool   `json:"locked"`
	} `json:"tags"`
	Open bool `json:"open"`
}

type EventRoom struct {
	RoomID        uuid.UUID `json:"roomId"`
	Place         string    `json:"place"`
	AllowTogether bool      `json:"sharedRoom"`
}

type EventRoomRes struct {
	RoomID        uuid.UUID `json:"roomId"`
	Place         string    `json:"place"`
	AllowTogether bool      `json:"sharedRoom"`
	Verified      bool      `json:"verified"`
}

type EventRoomDetail struct {
	AllowTogether bool `json:"sharedRoom"`
	RoomRes
}

type EventTagReq struct {
	Name string `json:"name"`
}

type EventScheduleStatusReq struct {
	Schedule ScheduleStatus `json:"schedule"`
}

// EventDetailRes is experimental
//go:generate gotypeconverter -s domain.Event -d EventDetailRes -o converter.go .
type EventDetailRes struct {
	ID          uuid.UUID          `json:"eventId"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Rooms       []EventRoomDetail  `json:"rooms"`
	Group       GroupRes           `json:"group"`
	GroupName   string             `json:"groupName" cvt:"Group"`
	TimeStart   time.Time          `json:"timeStart"`
	TimeEnd     time.Time          `json:"timeEnd"`
	CreatedBy   uuid.UUID          `json:"createdBy"`
	Admins      []uuid.UUID        `json:"admins"`
	Tags        []EventTagRes      `json:"tags"`
	Open        bool               `json:"open"`
	Attendees   []EventAttendeeRes `json:"attendees"`
	Model
}

type EventTagRes struct {
	ID     uuid.UUID `json:"tagId" cvt:"Tag"`
	Name   string    `json:"name" cvt:"Tag"`
	Locked bool      `json:"locked"`
}

type EventAttendeeRes struct {
	ID       uuid.UUID      `json:"userId" cvt:"UserID"`
	Schedule ScheduleStatus `json:"schedule"`
}

// EventRes is for multiple response
//go:generate gotypeconverter -s domain.Event -d EventRes -o converter.go .
//go:generate gotypeconverter -s []*domain.Event -d []EventRes -o converter.go .
type EventRes struct {
	ID          uuid.UUID          `json:"eventId"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Rooms       []EventRoomRes     `json:"rooms" cvt:"Room"`
	TimeStart   time.Time          `json:"timeStart"`
	TimeEnd     time.Time          `json:"timeEnd"`
	GroupID     uuid.UUID          `json:"groupId" cvt:"Group"`
	GroupName   string             `json:"groupName" cvt:"Group"`
	Admins      []uuid.UUID        `json:"admins"`
	Tags        []EventTagRes      `json:"tags"`
	CreatedBy   uuid.UUID          `json:"createdBy"`
	Open        bool               `json:"open"`
	Attendees   []EventAttendeeRes `json:"attendees"`
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
	// TODO
	// vevent.AddProperty("location", e.Room.Place)
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

func GenerateEventWebhookContent(method string, e *EventDetailRes, nofiticationTargets []string, origin string, isMention bool) string {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	timeFormat := "01/02(Mon) 15:04"
	var content string
	switch method {
	case http.MethodPost:
		content = "## イベントが作成されました" + "\n"
	case http.MethodPut:
		content = "## イベントが更新されました" + "\n"
	}
	content += fmt.Sprintf("### [%s](%s/events/%s)", e.Name, origin, e.ID) + "\n"
	content += fmt.Sprintf("- 主催: [%s](%s/groups/%s)", e.GroupName, origin, e.Group.ID) + "\n"
	content += fmt.Sprintf("- 日時: %s ~ %s", e.TimeStart.In(jst).Format(timeFormat), e.TimeEnd.In(jst).Format(timeFormat)) + "\n"
	content += fmt.Sprintf("- 場所: %s", e.Rooms[0].Place)
	for i := 1; i < len(e.Rooms); i++ {
		content += fmt.Sprintf(", %s", e.Rooms[i].Place)
	}
	content += "\n\n"

	if e.TimeStart.After(time.Now()) {
		content += "以下の方は参加予定の入力をお願いします:pray:" + "\n"
		prefix := "@"
		if !isMention {
			prefix = "@."
		}

		sort.Strings(nofiticationTargets)
		for _, nt := range nofiticationTargets {
			content += prefix + nt + " "
		}
		content += "\n\n\n"
	}

	// delete ">" if no description
	if strings.TrimSpace(e.Description) != "" {
		content += "> " + strings.ReplaceAll(e.Description, "\n", "\n> ")
	} else {
		content = strings.TrimRight(content, "\n")
	}

	return content
}
