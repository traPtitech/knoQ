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

const (
	AttendanceStampID = "93d376c3-80c9-4bb2-909b-2bbe2fbf9e93"
	AbsentStampID     = "544c04db-9cc3-4c0e-935d-571d4cf103a2"
	PendingStampID    = "bc9a3814-f185-4b3d-ac1f-3c8f12ad7b52"
)

func BotMessageStampsUpdatedHandler(p *payload.BotMessageStampsUpdated, gormRepo db.GormRepository) {
	post, err := gormRepo.GetPost(uuid.FromStringOrNil(p.MessageID))
	if err != nil {
		fmt.Println(err)
	}
	for _, stamp := range p.Stamps {
		var scheduleStatus domain.ScheduleStatus
		switch stamp.StampID {
		case AttendanceStampID:
			// 出席
			scheduleStatus = domain.Attendance
		case AbsentStampID:
			// 欠席
			scheduleStatus = domain.Absent
		case PendingStampID:
			// 未定
			scheduleStatus = domain.Pending
		}
		err := gormRepo.UpsertEventSchedule(post.EventID, uuid.FromStringOrNil(stamp.UserID), scheduleStatus)
		if err != nil {
			fmt.Println(err)
		}
	}
}
