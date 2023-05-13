package utils

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filter"
	"github.com/traPtitech/knoQ/infra/db"
)

// InitPostEventToTraQ 現在(job実行)から24時間以内に始まるイベントを取得し、
// webhookでtraQに送るjobを作成。
func InitPostEventToTraQ(repo *db.GormRepository, secret, channelID, webhookID, origin string) func() {
	job := func() {
		now := time.Now().AddDate(0, 0, 0)
		tomorrow := now.AddDate(0, 0, 1)

		rooms, _ := repo.GetAllRooms(now, tomorrow, uuid.Nil)
		events, _ := repo.GetAllEvents(filter.FilterTime(now, tomorrow))
		message := createMessage(now, rooms, events, origin)
		_ = RequestWebhook(message, secret, channelID, webhookID, 1)
	}

	return job
}

func createMessage(t time.Time, rooms []*domain.Room, events []*db.Event, origin string) string {
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
		} else {
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
