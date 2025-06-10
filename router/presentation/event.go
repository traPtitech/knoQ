package presentation

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/utils/tz"

	ics "github.com/arran4/golang-ical"
	"github.com/gofrs/uuid"
)

type ScheduleStatus int

const (
	Pending ScheduleStatus = iota + 1
	Attendance
	Absent
)

// EventReqWrite is
//
//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s EventReqWrite -d domain.WriteEventParams -o converter.go .
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
	Open bool `json:"open"`
}

type EventTagReq struct {
	Name string `json:"name"`
}

type EventScheduleStatusReq struct {
	Schedule ScheduleStatus `json:"schedule"`
}

// EventDetailRes is experimental
//
// //go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s domain.Event -d EventDetailRes -o converter.go .
type EventDetailRes struct {
	ID            uuid.UUID          `json:"eventId"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	Room          RoomRes            `json:"room"`
	Group         GroupRes           `json:"group"`
	Place         string             `json:"place" cvt:"Room"`
	GroupName     string             `json:"groupName" cvt:"Group"`
	TimeStart     time.Time          `json:"timeStart"`
	TimeEnd       time.Time          `json:"timeEnd"`
	CreatedBy     uuid.UUID          `json:"createdBy"`
	Admins        []uuid.UUID        `json:"admins"`
	Tags          []EventTagRes      `json:"tags"`
	AllowTogether bool               `json:"sharedRoom"`
	Open          bool               `json:"open"`
	Attendees     []EventAttendeeRes `json:"attendees"`
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
//
//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s domain.Event -d EventRes -o converter.go .
//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s []*domain.Event -d []EventRes -o converter.go .
type EventRes struct {
	ID            uuid.UUID          `json:"eventId"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	AllowTogether bool               `json:"sharedRoom"`
	TimeStart     time.Time          `json:"timeStart"`
	TimeEnd       time.Time          `json:"timeEnd"`
	RoomID        uuid.UUID          `json:"roomId" cvt:"Room"`
	GroupID       uuid.UUID          `json:"groupId" cvt:"Group"`
	Place         string             `json:"place" cvt:"Room"`
	GroupName     string             `json:"groupName" cvt:"Group"`
	Admins        []uuid.UUID        `json:"admins"`
	Tags          []EventTagRes      `json:"tags"`
	CreatedBy     uuid.UUID          `json:"createdBy"`
	Open          bool               `json:"open"`
	Attendees     []EventAttendeeRes `json:"attendees"`
	Model
}

func iCalVeventFormat(e *domain.Event, host string, userMap map[uuid.UUID]*domain.User) *ics.VEvent {
	vevent := ics.NewEvent(e.ID.String())
	vevent.SetDtStampTime(time.Now().UTC())
	vevent.SetStartAt(e.TimeStart.UTC())
	vevent.SetEndAt(e.TimeEnd.UTC())
	vevent.SetCreatedTime(e.CreatedAt.UTC())
	vevent.SetModifiedAt(e.UpdatedAt.UTC())
	vevent.SetSummary(e.Name)
	e.Description += "\n\n"
	e.Description += "-----------------------------------\n"
	e.Description += "イベント詳細ページ\n"
	e.Description += fmt.Sprintf("%s/events/%v", host, e.ID)
	vevent.SetDescription(e.Description)
	vevent.SetLocation(e.Room.Name)
	vevent.SetOrganizer(e.CreatedBy.DisplayName)
	for _, v := range e.Attendees {
		user, ok := userMap[v.UserID]
		if !ok {
			continue
		}

		userName := fmt.Sprintf("@%s", user.Name)
		userDisplayName := ics.WithCN(user.DisplayName)
		var ps ics.ParticipationStatus
		switch v.Schedule {
		case domain.Attendance:
			ps = ics.ParticipationStatusAccepted
		case domain.Absent:
			ps = ics.ParticipationStatusDeclined
		default:
			ps = ics.ParticipationStatusNeedsAction
		}
		vevent.AddAttendee(userName, ps, userDisplayName)
	}
	return vevent
}

func ICalFormat(events []*domain.Event, host string, userMap map[uuid.UUID]*domain.User) *ics.Calendar {
	var std ics.Standard
	std.ComponentBase.AddProperty(ics.ComponentProperty(ics.PropertyTzoffsetfrom), "+0900")
	std.ComponentBase.AddProperty(ics.ComponentProperty(ics.PropertyTzoffsetto), "+0900")
	std.ComponentBase.AddProperty(ics.ComponentProperty(ics.PropertyTzname), "JST")
	std.ComponentBase.AddProperty(ics.ComponentPropertyDtStart, "19700101T000000")

	tz := ics.NewTimezone("Asia/Tokyo")
	tz.Components = append(tz.Components, &std)

	cal := ics.NewCalendar()
	cal.AddVTimezone(tz)
	for _, e := range events {
		vevent := iCalVeventFormat(e, host, userMap)
		cal.AddVEvent(vevent)
	}
	return cal
}

func GenerateEventWebhookContent(method string, e *EventDetailRes, nofiticationTargets []string, origin string, isMention bool) string {
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
	content += fmt.Sprintf("- 日時: %s ~ %s", e.TimeStart.In(tz.JST).Format(timeFormat), e.TimeEnd.In(tz.JST).Format(timeFormat)) + "\n"
	content += fmt.Sprintf("- 場所: %s", e.Room.Place) + "\n"
	content += "\n"

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

func ConvdomainEventToEventDetailRes(src domain.Event) (dst EventDetailRes) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.Description = src.Description
	if src.IsRoomEvent {
		dst.Room = ConvdomainRoomToRoomRes(*src.Room)
		dst.Place = src.Room.Name
	} else {
		dst.Place = src.Venue.String
	}
	dst.Group = convdomainGroupToGroupRes(src.Group)
	dst.GroupName = src.Group.Name
	dst.TimeStart = src.TimeStart
	dst.TimeEnd = src.TimeEnd
	dst.CreatedBy = convdomainUserTouuidUUID(src.CreatedBy)
	dst.Admins = make([]uuid.UUID, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = convdomainUserTouuidUUID(src.Admins[i])
	}
	dst.Tags = make([]EventTagRes, len(src.Tags))
	for i := range src.Tags {
		dst.Tags[i] = convdomainEventTagToEventTagRes(src.Tags[i])
	}
	dst.AllowTogether = src.AllowTogether
	dst.Open = src.Open
	dst.Attendees = make([]EventAttendeeRes, len(src.Attendees))
	for i := range src.Attendees {
		dst.Attendees[i] = convdomainAttendeeToEventAttendeeRes(src.Attendees[i])
	}
	dst.Model = Model(src.Model)
	return
}
