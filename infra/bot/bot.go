package bot

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/traPtitech/traq-ws-bot/payload"
)

var Bot *traqwsbot.Bot

func BotMessageStampsUpdatedHandler(p *payload.BotMessageStampsUpdated, gormRepo db.GormRepository) {
	post, err := gormRepo.GetPost(uuid.FromStringOrNil(p.MessageID))
	if err != nil {
		fmt.Println(err)
	}
	for _, stamp := range p.Stamps {
		var scheduleStatus domain.ScheduleStatus
		switch stamp.StampID {
		case "8658c060-6f8e-46f6-8ee8-3fbc63f2ed4d":
			// 出席
			scheduleStatus = domain.Attendance
		case "7813a8d6-77ab-446d-af1c-559db2ccab10":
			// 欠席
			scheduleStatus = domain.Absent
		case "0d9c5a46-61bc-4c9c-873f-99eee45757fc":
			// 未定
			scheduleStatus = domain.Pending
		}
		err := gormRepo.UpsertEventSchedule(post.EventID, uuid.FromStringOrNil(stamp.UserID), scheduleStatus)
		if err != nil {
			fmt.Println(err)
		}
	}
}
