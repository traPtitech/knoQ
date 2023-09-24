package utils

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net/http"
	// "net/url"
	// "path"
	// "strconv"
	// "strings"

	// "github.com/labstack/echo/v4"
	"github.com/traPtitech/go-traq"
)

const baseURL = "https://q.trap.jp/api/v3"

// RequestWebhook q.trap/jp にメッセージを送信します。
func RequestWebhook(message, secret, channelID, webhookID string, embed int) error {
	// u, err := url.Parse(baseURL + "/webhooks")
	// if err != nil {
	// 	return err
	// }
	// u.Path = path.Join(u.Path, webhookID)
	// query := u.Query()
	// query.Set("embed", strconv.Itoa(embed))
	// u.RawQuery = query.Encode()

	// req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(message))
	// if err != nil {
	// 	return err
	// }
	// req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
	// req.Header.Set("X-TRAQ-Signature", calcSignature(message, secret))
	// if channelID != "" {
	// 	req.Header.Set("X-TRAQ-Channel-Id", channelID)
	// }

	// res, err := http.DefaultClient.Do(req)
	// if err != nil {
	// 	return err
	// }
	// if res.StatusCode >= 400 {
	// 	return errors.New(http.StatusText(res.StatusCode))
	// }

	// ここから go-traq 書き換え
	xTRAQSignature := calcSignature(message, secret)
	configuration := traq.NewConfiguration()
  apiClient := traq.NewAPIClient(configuration)

	res, err := apiClient.WebhookApi.PostWebhook(context.Background(),webhookID).XTRAQChannelId(channelID).XTRAQSignature(xTRAQSignature).Embed(int32(embed)).Body(message).Execute()
	if err != nil{
		return err
	}
	if res.StatusCode >= 400 {
		return errors.New(http.StatusText(res.StatusCode))
	}

	// ここまで go-traq 書き換え

	return nil
}

func calcSignature(message, secret string) string {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
