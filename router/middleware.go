package router

import (
	"errors"
	"fmt"
	log "room/logging"
	repo "room/repository"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"

	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
)

const requestUserStr string = "Request-User"

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
		id := c.Request().Header.Get("X-Showcase-User")
		if len(id) == 0 || id == "-" {
			// test用
			id = "c3f29c92-23d8-48f5-9553-002a932afeaf"
		}
		userID, err := uuid.FromString(id)
		if err != nil {
			return internalServerError()
		}
		user, err := repo.GetUser(userID)
		if err != nil {
			return internalServerError()
		}
		c.Set(requestUserStr, user)
		err = next(c)
		return err
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
