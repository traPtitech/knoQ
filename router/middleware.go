package router

import (
	"net/http"
	repo "room/repository"
	log "room/logging"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
)

const requestUserStr string = "Request-User"

// CreatedByGetter get created user
type CreatedByGetter interface {
	GetCreatedBy() (string, error)
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
				tmp.Runtime = c.Get("Error-Runtime").(error).Error()
				logger.Info("server error", zap.Object("field", tmp))
			case httpCode >= 400:
				tmp.Runtime = c.Get("Error-Runtime").(error).Error()
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
			id = "fuji"
		}
		user, err := repo.GetUser(id)
		if err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError) // データベースエラー
		}
		c.Set(requestUserStr, user)
		err = next(c)
		return err
	}
}

// AdminUserMiddleware 管理者ユーザーか判定するミドルウェア
func AdminUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestUser := GetRequestUser(c)

		// 判定
		if !requestUser.Admin {
			return echo.NewHTTPError(http.StatusForbidden) // 管理者ユーザーでは無いのでエラー
		}

		return next(c)
	}
}

// GroupCreatedUserMiddleware グループ作成ユーザーか判定するミドルウェア
func GroupCreatedUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestUser := GetRequestUser(c)
		g := new(repo.Group)
		var err error
		g.ID, err = strconv.Atoi(c.Param("groupid"))
		if err != nil {
			return notFound(message(err.Error()))
		}
		IsVerigy, err := VerifyCreatedUser(g, requestUser.TRAQID)
		if err != nil {
			return internalServerError()
		}
		if !IsVerigy {
			return badRequest(
					message("You are not user by whom this group is created."),
					specification("Only the created-user can edit."))
		}

		err = next(c)
		return err
	}
}

// EventCreatedUserMiddleware グループ作成ユーザーか判定するミドルウェア
func EventCreatedUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestUser := GetRequestUser(c)
		e := new(repo.Reservation)
		var err error
		e.ID, err = strconv.Atoi(c.Param("reservationid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		IsVerigy, err := VerifyCreatedUser(e, requestUser.TRAQID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		if !IsVerigy {
			return echo.NewHTTPError(http.StatusForbidden)
		}

		return next(c)
	}
}

// VerifyCreatedUser verify that request-user and created-user are the same
func VerifyCreatedUser(cbg CreatedByGetter, requestUser string) (bool, error) {
	createdByUser, err := cbg.GetCreatedBy()
	if err != nil {
		return false, err
	}
	if createdByUser != requestUser {
		return false, nil
	}
	return true, nil
}

// GetRequestUser リクエストユーザーを返します
func GetRequestUser(c echo.Context) repo.User {
	return c.Get(requestUserStr).(repo.User)
}
