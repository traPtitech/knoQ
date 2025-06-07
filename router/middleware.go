package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/router/logging"
	"github.com/traPtitech/knoQ/router/presentation"
	"github.com/traPtitech/knoQ/utils"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

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
			tmp := &logging.HTTPPayload{
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

// ServerVersionMiddleware X-KNOQ-VERSIONをレスポンスヘッダーを追加するミドルウェア
func ServerVersionMiddleware(version string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("X-KNOQ-VERSION", version)
			return next(c)
		}
	}
}

// TraQUserMiddleware traQユーザーか判定するミドルウェア
// TODO funcname fix
func (h *Handlers) TraQUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
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

// PrivilegeUserMiddleware 管理者ユーザーか判定するミドルウェア
func (h *Handlers) PrivilegeUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 判定
		if !h.Repo.IsPrivilege(getConinfo(c)) {
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
func (h *Handlers) WebhookEventHandler(c echo.Context, _, resBody []byte) {
	if c.Response().Status >= 400 {
		return
	}

	e := new(presentation.EventDetailRes)
	err := json.Unmarshal(resBody, e)
	if err != nil {
		return
	}

	users, err := h.Repo.GetAllUsers(false, true, getConinfo(c))
	if err != nil {
		return
	}
	usersMap := createUserMap(users)
	notificationTargets := make([]string, 0)

	if e.TimeEnd.Before(time.Now()) {
		return
	}

	// TODO fix: IDを環境変数などで定義すべき
	traPGroupID := uuid.Must(uuid.FromString("11111111-1111-1111-1111-111111111111"))
	if e.Group.ID == traPGroupID {
		groups, err := h.Repo.GetGradeGroupNames(getConinfo(c))
		if err != nil {
			h.Logger.Error("failed to get groups", zap.Error(err))
			return
		}

		notificationTargets = append(notificationTargets, groups...)
	} else {
		for _, attendee := range e.Attendees {
			if attendee.Schedule == presentation.Pending {
				user, ok := usersMap[attendee.ID]
				if ok {
					notificationTargets = append(notificationTargets, user.Name)
				}
			}
		}
	}

	content := presentation.GenerateEventWebhookContent(c.Request().Method, e, notificationTargets, h.Origin, !domain.DEVELOPMENT)

	_ = utils.RequestWebhook(content, h.WebhookSecret, h.ActivityChannelID, h.WebhookID, 1)
}

// getRequestUserID sessionからuserを返します
func getRequestUserID(c echo.Context) (uuid.UUID, error) {
	sess, err := session.Get("session", c)
	if err != nil {
		setMaxAgeMinus(c)
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

func setMaxAgeMinus(c echo.Context) {
	sess := &http.Cookie{
		Path:     "/",
		Name:     "session",
		HttpOnly: true,
		MaxAge:   -1,
	}
	c.SetCookie(sess)
}
