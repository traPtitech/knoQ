package router

import (
	"net/http"
	repo "room/repository"
	"room/utils"

	"github.com/labstack/echo/v4"
)

// HandleGetUserMe ヘッダー情報からuser情報を取得
// 認証状態を確認
func (h *Handlers) HandleGetUserMe(c echo.Context) error {
	// WIP
	token, _ := getRequestUserToken(c)
	UserGroupRepo := h.InitExternalUserGroupRepo(token, repo.V3)

	userID, _ := getRequestUserID(c)
	user, err := UserGroupRepo.GetUser(userID)
	if err != nil {
		if err.Error() == http.StatusText(http.StatusUnauthorized) {
			// 認証が切れている
			if err = deleteRequestUserToken(c); err != nil {
				return judgeErrorResponse(err)
			}
			return unauthorized(message("Your auth is expired"))
		}
		return internalServerError()
	}
	tmp, _ := h.Repo.GetUser(userID)
	user.Admin = tmp.Admin

	return c.JSON(http.StatusOK, user)
}

// HandleGetUsers ユーザーすべてを取得
func (h *Handlers) HandleGetUsers(c echo.Context) error {
	requestUserToken, _ := getRequestUserToken(c)
	users, err := utils.GetUsers(requestUserToken)
	if err != nil {
		if err.Error() == http.StatusText(http.StatusUnauthorized) {
			// 認証が切れている
			if err = deleteRequestUserToken(c); err != nil {
				return judgeErrorResponse(err)
			}
			return unauthorized(message("Your auth is expired"))
		}
		return internalServerError()
	}

	return c.JSON(http.StatusOK, users)
}
