package router

import (
	"errors"
	"fmt"
	log "room/logging"
	repo "room/repository"
	"room/utils"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"

	traQutils "github.com/traPtitech/traQ/utils"
)

const requestUserStr string = "Request-User"
const authScheme string = "Bearer"

var traQjson = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
	TagKey:                 "traq",
}

// CreatedByGetter get created user
type CreatedByGetter interface {
	GetCreatedBy() (uuid.UUID, error)
}

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
				errorRuntime, ok := c.Get("Error-Runtime").(error)
				if ok {
					tmp.ErrorLocation = errorRuntime.Error()
				} else {
					tmp.ErrorLocation = "no data"
				}
				logger.Info("server error", zap.Object("field", tmp))
			case httpCode >= 400:
				errorRuntime, ok := c.Get("Error-Runtime").(error)
				if ok {
					tmp.ErrorLocation = errorRuntime.Error()
				} else {
					tmp.ErrorLocation = "no data"
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
func TraQUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var user repo.User
		userSess := new(repo.UserSession)
		sess, err := session.Get("r_session", c)
		if err != nil {
			return unauthorized()
		}
		token, ok := sess.Values["token"].(string)
		if !ok {
			// token create
			token = traQutils.RandAlphabetAndNumberString(32)
			sess.Values["token"] = token
			sess.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   86400 * 7,
				HttpOnly: true,
			}
			// create DB record
			userSess.Token = token
			if err := userSess.Create(); err != nil {
				return internalServerError()
			}
			sess.Save(c.Request(), c.Response())
		}
		ah := c.Request().Header.Get(echo.HeaderAuthorization)
		if len(ah) > 0 {
			// AuthorizationヘッダーがあるためOAuth2で検証
			// Authorizationスキーム検証
			l := len(authScheme)
			if !(len(ah) > l+1 && ah[:l] == authScheme) {
				return unauthorized(message("invalid authorization scheme"))
			}

			// OAuth2 Token検証
			// Todo traQ /users/me
			body, err := utils.GetUserMe(ah)
			if err != nil {
				return unauthorized(message(err.Error()))
			}
			err = traQjson.Froze().Unmarshal(body, &user)
			if err != nil {
				return internalServerError()
			}

			// Todo session を認証状態にする
			// DB update
			userSess = &repo.UserSession{
				Token:         token,
				UserID:        user.ID,
				Authorization: ah,
			}
			if err := userSess.Update(); err != nil {
				return unauthorized()
			}

		} else {
			// take from DB
			userSess.Token = token
			if err := userSess.Get(); err != nil {
				return unauthorized()
			}
			// 認証されてないなら空文字列
			if userSess.Authorization == "" {
				return unauthorized()
			}
		}
		user.ID = userSess.UserID
		user.Auth = userSess.Authorization
		err = repo.DB.FirstOrCreate(&user).Error
		if err != nil {
			return internalServerError()
		}
		c.Set(requestUserStr, user)
		return next(c)
	}
}

// AdminUserMiddleware 管理者ユーザーか判定するミドルウェア
func AdminUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestUser := getRequestUser(c)

		// 判定
		if !requestUser.Admin {
			return forbidden(
				message("You are not admin user."),
				specification("Only admin user can request."),
			)
		}

		return next(c)
	}
}

// GroupCreatedUserMiddleware グループ作成ユーザーか判定するミドルウェア
func GroupCreatedUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestUser := getRequestUser(c)
		g := new(repo.Group)
		var err error
		g.ID, err = getRequestGroupID(c)
		if err != nil || g.ID == uuid.Nil {
			internalServerError()
		}
		IsVerigy, err := verifyCreatedUser(g, requestUser.ID)
		if err != nil {
			return internalServerError()
		}
		if !IsVerigy {
			return badRequest(
				message("You are not user by whom this group is created."),
				specification("Only the author can request."),
			)
		}

		return next(c)
	}
}

// EventCreatedUserMiddleware グループ作成ユーザーか判定するミドルウェア
func EventCreatedUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestUser := getRequestUser(c)
		event := new(repo.Event)
		var err error
		event.ID, err = getRequestEventID(c)
		if err != nil {
			return internalServerError()
		}

		IsVerigy, err := verifyCreatedUser(event, requestUser.ID)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return notFound(message(fmt.Sprintf("EventID: %v does not exist.", c.Param("eventid"))))
			}
			return internalServerError()
		}
		if !IsVerigy {
			return badRequest(
				message("You are not user by whom this even is created."),
				specification("Only the author can request."),
			)
		}

		return next(c)
	}
}

func EventIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		event := new(repo.Event)
		var err error
		event.ID, err = uuid.FromString(c.Param("eventid"))
		if err != nil || event.ID == uuid.Nil {
			return notFound(message(fmt.Sprintf("EventID: %v does not exist.", c.Param("eventid"))))
		}
		if err := repo.DB.Select("id").First(&event).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return notFound(message(fmt.Sprintf("EventID: %v does not exist.", c.Param("eventid"))))
			}
			return internalServerError()
		}
		c.Set("EventID", event.ID)

		return next(c)
	}
}

func GroupIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		group := new(repo.Group)
		var err error
		group.ID, err = uuid.FromString(c.Param("groupid"))
		if err != nil || group.ID == uuid.Nil {
			return notFound(message(fmt.Sprintf("GroupID: %v does not exist.", c.Param("groupid"))))
		}
		if err := repo.DB.Select("id").First(&group).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return notFound(message(fmt.Sprintf("GroupID: %v does not exist.", c.Param("groupid"))))
			}
			return internalServerError()
		}
		c.Set("GroupID", group.ID)

		return next(c)

	}
}

// verifyCreatedUser verify that request-user and created-user are the same
func verifyCreatedUser(cbg CreatedByGetter, requestUser uuid.UUID) (bool, error) {
	createdByUser, err := cbg.GetCreatedBy()
	if err != nil {
		return false, err
	}
	if createdByUser != requestUser {
		return false, nil
	}
	return true, nil
}

// getRequestUser リクエストユーザーを返します
func getRequestUser(c echo.Context) repo.User {
	return c.Get(requestUserStr).(repo.User)
}

// getRequestEventID :eventidを返します
func getRequestEventID(c echo.Context) (uuid.UUID, error) {
	eventID, ok := c.Get("EventID").(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("EventID is not set")
	}
	return eventID, nil
}

// getRequestGroupID :groupidを返します
func getRequestGroupID(c echo.Context) (uuid.UUID, error) {
	groupID, ok := c.Get("GroupID").(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("GroupID is not set")
	}
	return groupID, nil
}
