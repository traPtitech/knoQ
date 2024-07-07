package utils

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/go-traq"
	"github.com/traPtitech/knoQ/infra/bot"
)

// RequestBotPost q.trap/jp にメッセージを送信します。
func RequestBotPost(messageText, channelID string) (uuid.UUID, error) {
	message, _, err := bot.Bot.API().
		MessageApi.
		PostMessage(context.Background(), channelID).
		PostMessageRequest(traq.PostMessageRequest{
			Content: messageText,
		}).
		Execute()
	return uuid.FromStringOrNil(message.Id), err
}

// RequestBotStatusStamp :white_check_mark:, :no_entry_sign:, :loading: のスタンプを送信します。
func RequestBotStatusStamp(messageID uuid.UUID) error {
	stamps := []string{bot.AttendanceStampID, bot.AbsentStampID, bot.PendingStampID}
	for _, stampID := range stamps {
		_, err := bot.Bot.API().
			MessageApi.AddMessageStamp(context.Background(), messageID.String(), stampID).Execute()
		if err != nil {
			return err
		}
	}
	return nil
}
