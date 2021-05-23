package utils

import (
	"time"

	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filter"
	"github.com/traPtitech/knoQ/presentation"
)

// InitPostEventToTraQ 現在(job実行)から24時間以内に始まるイベントを取得し、
// webhookでtraQに送るjobを作成。
func InitPostEventToTraQ(repo interface {
	domain.EventRepository
	domain.RoomRepository
}, secret, channelID, webhookID, origin string) func() {
	job := func() {
		now := time.Now().AddDate(0, 0, 0)
		tomorrow := now.AddDate(0, 0, 1)

		rooms, _ := repo.GetAllRooms(now, tomorrow)
		events, _ := repo.GetEvents(filter.FilterTime(now, tomorrow), nil)
		message := presentation.SchedulerMessageFormat(now, rooms, events, origin)
		RequestWebhook(message, secret, channelID, webhookID, 1)
	}

	return job
}

func InitRemindEvent(repo domain.EventRepository, secret, channelID, webhookID, origin string) func() {
	job := func() {
		now := time.Now()
		events, _ := repo.GetEvents(filter.FilterTime(now, now.Add(1*time.Minute)), nil)
		for _, event := range events {
			message := "## 【リマインド】\n"
			content := presentation.WebhookMessageFormat(
				presentation.ConvdomainEventToEventDetailRes(*event),
				"", origin)
			message += content
			RequestWebhook(message, secret, channelID, webhookID, 1)
		}
	}

	return job
}
