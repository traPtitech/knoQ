package utils

import (
	"context"

	"github.com/traPtitech/go-traq"
	"github.com/traPtitech/knoQ/infra/bot"
)

// RequestBotPost q.trap/jp にメッセージを送信します。
func RequestBotPost(message, channelID string) error {
	_, _, err := bot.Bot.API().
		MessageApi.
		PostMessage(context.Background(), channelID).
		PostMessageRequest(traq.PostMessageRequest{
			Content: message,
		}).
		Execute()
	return err
}
