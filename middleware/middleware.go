package middleware

import (
	"net/http"
	repo "room/repository"
	"strconv"

	"github.com/labstack/echo"
)

const traQID = "traQID"

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
		c.Set("Request-User", user)
		return next(c)
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

// CreatedByGetter get created user
type CreatedByGetter interface {
	GetCreatedBy() (string, error)
}

// GroupCreatedUserMiddleware グループ作成ユーザーか判定するミドルウェア
func GroupCreatedUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestUser := GetRequestUser(c)
		g := new(repo.Group)
		var err error
		g.ID, err = strconv.Atoi(c.Param("groupid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		IsVerigy, err := VerifyCreatedUser(g, requestUser.TRAQID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		if !IsVerigy {
			return echo.NewHTTPError(http.StatusForbidden)
		}

		return next(c)
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

// GetRequestUser リクエストユーザーのtraQIDを返します
func GetRequestUser(c echo.Context) repo.User {
	return c.Get("Request-User").(repo.User)
}
