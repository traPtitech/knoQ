package utils

import (
	"fmt"
	repo "room/repository"
	"time"
)

// InitPostEventToTraQ 現在(job実行)から24時間以内に始まるイベントを取得し、
// webhookでtraQに送るjobを作成。
func InitPostEventToTraQ(repo interface {
	repo.EventRepository
	repo.RoomRepository
}, secret, channelID, webhookID, origin string) func() {
	job := func() {
		now := time.Now().AddDate(0, 0, 0)
		tomorrow := now.AddDate(0, 0, 1)
		rooms, _ := repo.GetAllRooms(&now, &tomorrow)
		message := createMessage(now, rooms, origin)
		// RequestWebhook(message, secret, channelID, webhookID, 1)
		fmt.Println(message)
	}

	return job
}

func createMessage(t time.Time, rooms []*repo.Room, origin string) (message string) {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	date := t.In(jst).Format("01/02(Mon)")
	public := map[bool]string{
		true:  "**(traP講義室)**",
		false: "",
	}
	combined := map[bool]string{
		true:  "併用可",
		false: "併用不可",
	}
	if rooms == nil {
		return fmt.Sprintf("%sの進捗部屋は、予約を取っていないようです。\n", date)
	}
	message = fmt.Sprintf("%sの進捗部屋は、\n", date)
	for _, room := range rooms {
		message += fmt.Sprintf("- %s %s %s ~ %s\n", public[room.Public], room.Place,
			room.TimeStart.In(jst).Format("15:04"), room.TimeEnd.In(jst).Format("15:04"))
		for _, event := range room.Events {
			message += fmt.Sprintf("\t- %s ~ [%s](%s/events/%s) (%s)\n", event.TimeStart.In(jst).Format("15:04"),
				event.Name, origin, event.ID, combined[event.AllowTogether])
		}
	}
	return
}
