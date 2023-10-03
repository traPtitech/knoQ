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

var jst, _ = time.LoadLocation("Asia/Tokyo")

type timeTable struct {
	name           string
	start          time.Time
	displayDefault bool
}

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

func setTimeFromString(t time.Time, str string) time.Time {
	s, _ := time.Parse(time.TimeOnly, str)
	return time.Date(t.Year(), t.Month(), t.Day(), s.Hour(), s.Minute(), s.Second(), 0, jst)
}

// makeRoomAvailableByTimeTable timeTables の各時間帯を行、rooms の各部屋を列とする表を map 形式で作成する。 unVerified の部屋は無視する。
func makeRoomAvailableByTimeTable(rooms []*domain.Room, timeTables []timeTable, date time.Time) []map[string]string {
	roomAvailable := make([]map[string]string, len(timeTables))
	for i := range roomAvailable {
		roomAvailable[i] = make(map[string]string)
	}
	for _, room := range rooms {
		if !room.Verified {
			continue
		}

		ts, te := room.TimeStart, room.TimeEnd
		for i, row := range timeTables {
			rowNextStart := setTimeFromString(date, "23:59:59")
			if i < len(timeTables)-1 {
				rowNextStart = timeTables[i+1].start
			}

			rs := row.start
			// 進捗部屋使用開始 <= n限開始 < 進捗部屋使用終了
			if (ts.Before(rs) || ts.Equal(rs)) && rs.Before(te) {
				if rowNextStart.Before(te) || rowNextStart.Equal(te) {
					// n限の間全使用
					roomAvailable[i][room.Place] = ":white_check_mark:"
				} else {
					// n限の途中で使用終了
					roomAvailable[i][room.Place] = fmt.Sprintf("- %s", te.Format("15:04"))
				}
				continue
			}

			// n限開始 < 進捗部屋使用開始 < n+1限開始
			if rs.Before(ts) && ts.Before(rowNextStart) {
				if rowNextStart.Before(te) || rowNextStart.Equal(te) {
					// n限の途中で使用開始し、n限の間は全使用
					roomAvailable[i][room.Place] = fmt.Sprintf("%s -", ts.Format("15:04"))
				} else {
					// n限の途中で使用開始し、n限の途中で使用終了
					roomAvailable[i][room.Place] = fmt.Sprintf("%s - %s", ts.Format("15:04"), te.Format("15:04"))
				}
				continue
			}

			// n限の間は進捗部屋を使用しない
			if _, ok := roomAvailable[i][room.Place]; !ok {
				roomAvailable[i][room.Place] = ":regional_indicator_null:"
			}
		}
	}
	return roomAvailable
}

func createMessage(t time.Time, rooms []*domain.Room, events []*db.Event, origin string) string {
	date := t.In(jst).Format("01/02(Mon)")
	combined := map[bool]string{
		true:  "(併用可)",
		false: "",
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
			timeTables := []timeTable{
				{":sunny:", setTimeFromString(t, "00:00:00"), false},
				{"1-2", setTimeFromString(t, "08:50:00"), true},
				{"3-4", setTimeFromString(t, "10:45:00"), true},
				{"昼", setTimeFromString(t, "12:25:00"), true},
				{"5-6", setTimeFromString(t, "13:30:00"), true},
				{"7-8", setTimeFromString(t, "15:25:00"), true},
				{"9-10", setTimeFromString(t, "17:15:00"), true},
				{":crescent_moon:", setTimeFromString(t, "18:55:00"), false},
			}
			roomAvailable := makeRoomAvailableByTimeTable(rooms, timeTables, t)

			roomMessage = fmt.Sprintf(
				"| time | %s |\n| :---: | %s \n",
				strings.Join(verifiedRoomNames, " | "),
				strings.Repeat(" :---: |", len(verifiedRoomNames)),
			)
			for i, row := range timeTables {

				forceDisplay := slices.ContainsFunc(verifiedRoomNames, func(vr string) bool {
					return roomAvailable[i][vr] != ":regional_indicator_null:"
				})

				if !row.displayDefault && !forceDisplay {
					continue
				}

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
