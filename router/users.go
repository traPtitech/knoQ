package router

import (
	"net/http"
	repo "room/repository"
	"room/utils"

	"github.com/labstack/echo/v4"
)

// HandleGetUserMe ヘッダー情報からuser情報を取得
// 認証状態を確認
func HandleGetUserMe(c echo.Context) error {
	requestUser := getRequestUser(c)
	_, err := utils.GetUserMe(requestUser.Auth)
	if err != nil {
		if err.Error() == http.StatusText(http.StatusUnauthorized) {
			// 認証が切れている
			if err = repo.DeleteAuth(requestUser.Auth); err != nil {
				return judgeErrorResponse(err)
			}
			return unauthorized(message("Your auth is expired"))
		}
		return internalServerError()
	}

	return c.JSON(http.StatusOK, requestUser)
}

// HandleGetUsers ユーザーすべてを取得
func HandleGetUsers(c echo.Context) error {
	users := []repo.User{}
	if err := repo.DB.Find(&users).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, users)
}
