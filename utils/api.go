package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

const baseURL = "https://q.trap.jp/api/v3"

func GetUserMe(token string) ([]byte, error) {
	return APIGetRequest(token, "/users/me")
}

func GetUsers(token string) ([]byte, error) {
	return APIGetRequest(token, "/users")
}

func APIGetRequest(token, endpoint string) ([]byte, error) {
	if token == "" {
		return nil, errors.New(http.StatusText(http.StatusUnauthorized))
	}
	req, err := http.NewRequest(http.MethodGet, baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 300 {
		return nil, errors.New(http.StatusText(res.StatusCode))
	}
	return ioutil.ReadAll(res.Body)
}

// RequestWebhook q.trap/jp にメッセージを送信します。
func RequestWebhook(message, secret, channelID, webhookID string, embed int) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, webhookID)
	query := u.Query()
	query.Set("embed", strconv.Itoa(embed))
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(message))
	if err != nil {
		return err
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
	req.Header.Set("X-TRAQ-Signature", calcSignature(message, secret))
	if channelID != "" {
		req.Header.Set("X-TRAQ-Channel-Id", channelID)
	}

	res, err := http.DefaultClient.Do(req)
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
