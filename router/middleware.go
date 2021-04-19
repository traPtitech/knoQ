package router

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	traQrandom "github.com/traPtitech/traQ/utils/random"

	log "github.com/traPtitech/knoQ/logging"
	"github.com/traPtitech/knoQ/presentation"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const requestUserStr string = "Request-User"
const authScheme string = "Bearer"

func AccessLoggingMiddleware(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			if err := next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()

			req := c.Request()
			res := c.Response()
			tmp := &log.HTTPPayload{
				RequestMethod: req.Method,
				Status:        res.Status,
				UserAgent:     req.UserAgent(),
				RemoteIP:      c.RealIP(),
				Referer:       req.Referer(),
				Protocol:      req.Proto,
				RequestURL:    req.URL.String(),
				RequestSize:   req.Header.Get(echo.HeaderContentLength),
				ResponseSize:  strconv.FormatInt(res.Size, 10),
				Latency:       strconv.FormatFloat(stop.Sub(start).Seconds(), 'f', 9, 64) + "s",
			}
			httpCode := res.Status
			switch {
			case httpCode >= 500:
				errorRuntime, ok := c.Get("Error").(error)
				if ok {
					tmp.Error = errorRuntime.Error()
				} else {
					tmp.Error = "no data"
				}
				logger.Info("server error", zap.Object("field", tmp))
			case httpCode >= 400:
				errorRuntime, ok := c.Get("Error").(error)
				if ok {
					tmp.Error = errorRuntime.Error()
				} else {
					tmp.Error = "no data"
				}
				logger.Info("client error", zap.Object("field", tmp))
			case httpCode >= 300:
				logger.Info("redirect", zap.Object("field", tmp))
			case httpCode >= 200:
				logger.Info("success", zap.Object("field", tmp))
			}
			return nil
		}
	}
}

// TraQUserMiddleware traQユーザーか判定するミドルウェア
// TODO funcname fix
func (h *Handlers) TraQUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return unauthorized(err)
		}
		_, ok := sess.Values["ID"].(string)
		if !ok {
			sess.Options = &h.SessionOption
			sess.Values["ID"] = traQrandom.SecureAlphaNumeric(10)
			sess.Save(c.Request(), c.Response())
			return unauthorized(err, needAuthorization(true))
		}
		userID, err := getRequestUserID(c)
		if err != nil || userID == uuid.Nil {
			return unauthorized(err, needAuthorization(true))
		}

		user, err := h.Repo.GetUserMe(getConinfo(c))
		if err != nil {
			return internalServerError(err)
		}

		// state check
		if user.State != 1 {
			return forbidden(errors.New("invalid user"))
		}
		return next(c)
	}
}

// PrevilegeUserMiddleware 管理者ユーザーか判定するミドルウェア
func (h *Handlers) PrevilegeUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 判定
		if !h.Repo.IsPrevilege(getConinfo(c)) {
			return forbidden(
				errors.New("not admin"),
				message("You are not admin user."),
				specification("Only admin user can request."),
			)
		}

		return next(c)
	}
}

// GroupAdminsMiddleware グループ管理ユーザーか判定するミドルウェア
func (h *Handlers) GroupAdminsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		groupID, err := getPathGroupID(c)
		if err != nil {
			return notFound(err)
		}
		if !h.Repo.IsGroupAdmins(groupID, getConinfo(c)) {
			return forbidden(
				errors.New("not createdBy"),
				message("You are not user by whom this group is created."),
				specification("Only the author can request."),
			)
		}
		return next(c)
	}
}

// EventAdminsMiddleware イベント管理ユーザーか判定するミドルウェア
func (h *Handlers) EventAdminsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		eventID, err := getPathEventID(c)
		if err != nil {
			return notFound(err)
		}

		if !h.Repo.IsEventAdmins(eventID, getConinfo(c)) {
			return forbidden(
				errors.New("not createdBy"),
				message("You are not user by whom this even is created."),
				specification("Only the author can request."),
			)
		}

		return next(c)
	}
}

// RoomAdminsMiddleware 部屋管理ユーザーか判定するミドルウェア
func (h *Handlers) RoomAdminsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		roomID, err := getPathRoomID(c)
		if err != nil {
			return notFound(err)
		}

		if !h.Repo.IsRoomAdmins(roomID, getConinfo(c)) {
			return forbidden(
				errors.New("not createdBy"),
				message("You are not user by whom this even is created."),
				specification("Only the author can request."),
			)
		}

		return next(c)
	}
}

// WebhookEventHandler is used with middleware.BodyDump
func (h *Handlers) WebhookEventHandler(c echo.Context, reqBody, resBody []byte) {
	if c.Response().Status >= 400 {
		return
	}

	e := new(presentation.EventDetailRes)
	err := json.Unmarshal(resBody, e)
	if err != nil {
		return
	}

	jst, _ := time.LoadLocation("Asia/Tokyo")
	timeFormat := "01/02(Mon) 15:04"
	var content string
	if c.Request().Method == http.MethodPost {
		content = "## イベントが作成されました" + "\n"
	} else if c.Request().Method == http.MethodPut {
		content = "## イベントが更新されました" + "\n"
	}
	content += fmt.Sprintf("### [%s](%s/events/%s)", e.Name, h.Origin, e.ID) + "\n"
	content += fmt.Sprintf("- 主催: [%s](%s/groups/%s)", e.GroupName, h.Origin, e.Group.ID) + "\n"
	content += fmt.Sprintf("- 日時: %s ~ %s", e.TimeStart.In(jst).Format(timeFormat), e.TimeEnd.In(jst).Format(timeFormat)) + "\n"
	content += fmt.Sprintf("- 場所: %s", e.Room.Place) + "\n"
	content += "\n"
	content += e.Description

	_ = RequestWebhook(content, h.WebhookSecret, h.ActivityChannelID, h.WebhookID, 1)
}

func RequestWebhook(message, secret, channelID, webhookID string, embed int) error {
	u, err := url.Parse("https://q.trap.jp/api/1.0/webhooks")
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

// getRequestUserID sessionからuserを返します
func getRequestUserID(c echo.Context) (uuid.UUID, error) {
	sess, err := session.Get("session", c)
	if err != nil {
		return uuid.Nil, err
	}
	userID, _ := sess.Values["userID"].(string)
	return uuid.FromString(userID)
}

// getPathEventID :eventidを返します
func getPathEventID(c echo.Context) (uuid.UUID, error) {
	eventID, err := uuid.FromString(c.Param("eventid"))
	if err != nil {
		return uuid.Nil, errors.New("EventID is not uuid")
	}
	return eventID, nil
}

// getPathGroupID :groupidを返します
func getPathGroupID(c echo.Context) (uuid.UUID, error) {
	groupID, err := uuid.FromString(c.Param("groupid"))
	if err != nil {
		return uuid.Nil, errors.New("GroupID is not uuid")
	}
	return groupID, nil
}

// getPathRoomID :roomidを返します
func getPathRoomID(c echo.Context) (uuid.UUID, error) {
	roomID, err := uuid.FromString(c.Param("roomid"))
	if err != nil {
		return uuid.Nil, errors.New("RoomID is not uuid")
	}
	return roomID, nil
}

// getPathUserID :useridを返します
func getPathUserID(c echo.Context) (uuid.UUID, error) {
	userID, err := uuid.FromString(c.Param("userid"))
	if err != nil {
		return uuid.Nil, errors.New("UserID is not uuid")
	}
	return userID, nil
}
