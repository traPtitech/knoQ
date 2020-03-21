package router

import (
	"net/http"
	repo "room/repository"

	"github.com/labstack/echo/v4"
)

// HandleGetUserMe ヘッダー情報からuser情報を取得
// 認証状態を確認
func (h *Handlers) HandleGetUserMe(c echo.Context) error {
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

	return c.JSON(http.StatusOK, formatUserRes(user))
}

// HandleGetUsers ユーザーすべてを取得
func (h *Handlers) HandleGetUsers(c echo.Context) error {
	token, _ := getRequestUserToken(c)
	UserGroupRepo := h.InitExternalUserGroupRepo(token, repo.V3)

	users, err := UserGroupRepo.GetAllUsers()
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
	gormUsers, err := h.Repo.GetAllUsers()
	if err != nil {
		return internalServerError()
	}
	// add admin field
	for _, user := range gormUsers {
		for i, u := range users {
			if user.ID == u.ID {
				users[i].Admin = user.Admin
			}
		}
	}
	res := make([]*UserRes, len(users))
	for i, u := range users {
		res[i] = formatUserRes(u)
	}
	return c.JSON(http.StatusOK, res)
}
