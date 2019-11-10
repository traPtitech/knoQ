package router

import (
	"net/http"

	"github.com/labstack/echo"
)

// HandleGetUserMe ヘッダー情報からuser情報を取得
func HandleGetUserMe(c echo.Context) error {
	traQID := getRequestUser(c)
	user, err := getUser(traQID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, user)
}

// HandleGetUsers ユーザーすべてを取得
func HandleGetUsers(c echo.Context) error {
	users := []User{}
	if err := db.Find(&users).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, users)
}
