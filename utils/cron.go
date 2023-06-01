package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filter"
	"github.com/traPtitech/knoQ/infra/db"
	"golang.org/x/exp/slices"
)

// InitPostEventToTraQ 現在(job実行)から24時間以内に始まるイベントを取得し、
// webhookでtraQに送るjobを作成。
func InitPostEventToTraQ(repo *db.GormRepository, secret, channelID, webhookID, origin string) func() {
	job := func() {
		now := setTimeFromString(time.Now().AddDate(0, 0, 0), "06:00:00")
		tomorrow := now.AddDate(0, 0, 1)

		rooms, _ := repo.GetAllRooms(now, tomorrow, uuid.Nil)
		events, _ := repo.GetAllEvents(filter.FilterTime(now, tomorrow))
		message := createMessage(now, rooms, events, origin)
		_ = RequestWebhook(message, secret, channelID, webhookID, 1)
	}

	return job
}

type timeTable struct {
	name  string
	start time.Time
	end   time.Time
}

func setTimeFromString(t time.Time, str string) time.Time {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	s, _ := time.Parse(time.TimeOnly, str)
	return time.Date(t.Year(), t.Month(), t.Day(), s.Hour(), s.Minute(), s.Second(), 0, jst)
}

// t1 <= t2
func timeLessThanOrEqual(t1, t2 time.Time) bool {
	return t1.Before(t2) || t1.Equal(t2)
}

func createMessage(t time.Time, rooms []*domain.Room, events []*db.Event, origin string) string {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	date := t.In(jst).Format("01/02(Mon)")
	combined := map[bool]string{
		true:  "(併用可)",
		false: "",
	}

	timeTables := []timeTable{
		{
			name:  "1-2",
			start: setTimeFromString(t, "08:50:00"),
			end:   setTimeFromString(t, "10:30:00"),
		}, {
			name:  "3-4",
			start: setTimeFromString(t, "10:45:00"),
			end:   setTimeFromString(t, "12:25:00"),
		}, {
			name:  "昼",
			start: setTimeFromString(t, "12:25:00"),
			end:   setTimeFromString(t, "13:45:00"),
		}, {
			name:  "5-6",
			start: setTimeFromString(t, "13:45:00"),
			end:   setTimeFromString(t, "15:25:00"),
		}, {
			name:  "7-8",
			start: setTimeFromString(t, "15:40:00"),
			end:   setTimeFromString(t, "17:20:00"),
		}, {
			name:  "9-10",
			start: setTimeFromString(t, "17:30:00"),
			end:   setTimeFromString(t, "19:10:00"),
		},
	}

	roomMessage := ""
	eventMessage := "本日開催されるイベントは、\n"

	var verifiedRoomNames []string

	if len(rooms) == 0 {
		roomMessage = fmt.Sprintf("%sの進捗部屋は、予約を取っていないようです。\n", date)
	} else {
		for _, room := range rooms {
			if room.Verified && !slices.Contains(verifiedRoomNames, room.Place) {
				verifiedRoomNames = append(verifiedRoomNames, room.Place)
			}
		}

		if len(verifiedRoomNames) == 0 {
			roomMessage = fmt.Sprintf("%sの進捗部屋は、予約を取っていないようです。\n", date)
		} else {
			roomAvailable := make([]map[string]string, len(timeTables))
			for i := range roomAvailable {
				roomAvailable[i] = make(map[string]string)
			}
			for _, room := range rooms {
				if !room.Verified {
					continue
				}

				for i, row := range timeTables {
					if timeLessThanOrEqual(room.TimeStart, row.start) { // room.TimeStart <= row.start
						switch {
						case timeLessThanOrEqual(row.end, room.TimeEnd): // row.end <= room.TimeEnd
							roomAvailable[i][room.Place] = ":white_check_mark:"

						case room.TimeEnd.After(row.start): // row.start < room.TimeEnd < row.end
							roomAvailable[i][room.Place] = fmt.Sprintf(" - %s", room.TimeEnd.Format("15:04"))

						default: // room.TimeEnd == row.start
							if _, ok := roomAvailable[i][room.Place]; !ok {
								roomAvailable[i][room.Place] = ":regional_indicator_null:"
							}
						}
					} else { // row.start < room.TimeStart
						if timeLessThanOrEqual(row.end, room.TimeStart) { // row.end <= room.TimeStart
							if _, ok := roomAvailable[i][room.Place]; !ok {
								roomAvailable[i][room.Place] = ":regional_indicator_null:"
							}
							continue
						}
						// row.start < room.TimeStart < row.end

						if timeLessThanOrEqual(row.end, room.TimeEnd) { // row.end <= room.TimeEnd
							roomAvailable[i][room.Place] = fmt.Sprintf("%s -", room.TimeStart.Format("15:04"))
							continue
						}

						// row.start < room.TimeEnd < row.end
						roomAvailable[i][room.Place] = fmt.Sprintf("%s - %s", room.TimeStart.Format("15:04"), room.TimeEnd.Format("15:04"))
					}
				}
			}

			roomMessage = fmt.Sprintf("| time | %s |\n| :---: | %s \n", strings.Join(verifiedRoomNames, " | "), strings.Repeat(" :---: |", len(verifiedRoomNames)))
			for i, row := range timeTables {
				roomMessage += fmt.Sprintf("| %s |", row.name)

				for _, col := range verifiedRoomNames {
					roomMessage += fmt.Sprintf(" %s |", roomAvailable[i][col])
				}
				roomMessage += "\n"
			}
			roomMessage += "\n"
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
