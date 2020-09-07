package router

import (
	"net/http"
	"room/router/service"

	"github.com/labstack/echo/v4"
	traQutils "github.com/traPtitech/traQ/utils"
)

// HandleGetUserMe ヘッダー情報からuser情報を取得
// 認証状態を確認
func (h *Handlers) HandleGetUserMe(c echo.Context) error {
	token, _ := getRequestUserToken(c)
	userID, _ := getRequestUserID(c)

	user, err := h.Dao.GetUser(token, userID)
	if err != nil {
		if err.Error() == http.StatusText(http.StatusUnauthorized) {
			h.Repo.ReplaceToken(userID, "")
			return forbidden(err, message("token is invalid."), needAuthorization(true))
		}
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, service.FormatUserRes(user))
}

// HandleGetUsers ユーザーすべてを取得
func (h *Handlers) HandleGetUsers(c echo.Context) error {
	token, _ := getRequestUserToken(c)
	userID, _ := getRequestUserID(c)

	users, err := h.Dao.GetAllUsers(token)
	if err != nil {
		if err.Error() == http.StatusText(http.StatusUnauthorized) {
			h.Repo.ReplaceToken(userID, "")
			return forbidden(err, message("token is invalid."), needAuthorization(true))
		}
		return judgeErrorResponse(err)
	}

	return c.JSON(http.StatusOK, service.FormatUsersRes(users))
}

func (h *Handlers) HandleUpdateiCal(c echo.Context) error {
	userID, _ := getRequestUserID(c)
	secret := traQutils.RandAlphabetAndNumberString(16)
	if err := h.Dao.Repo.UpdateiCalSecretUser(userID, secret); err != nil {
		return judgeErrorResponse(err)
	}
	return c.String(http.StatusOK, secret)
}
