package utils

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/traPtitech/go-traq"
)

// RequestWebhook q.trap/jp にメッセージを送信します。
func RequestWebhook(message, secret, channelID, webhookID string, embed int) error {
	configuration := traq.NewConfiguration()
	apiClient := traq.NewAPIClient(configuration)

	xTRAQSignature := calcSignature(message, secret)
	res, err := apiClient.WebhookApi.PostWebhook(context.TODO(), webhookID).
		XTRAQChannelId(channelID).XTRAQSignature(xTRAQSignature).
		Embed(int32(embed)).Body(message).Execute()
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return errors.New(http.StatusText(res.StatusCode))
	}

	return nil
}

func calcSignature(message, secret string) string {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
