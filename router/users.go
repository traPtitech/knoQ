package router

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/presentation"
	"github.com/traPtitech/knoQ/usecase/production"

	"github.com/labstack/echo/v4"
)

// HandleGetUserMe ヘッダー情報からuser情報を取得
// 認証状態を確認
func (h *Handlers) HandleGetUserMe(c echo.Context) error {
	user, err := h.Repo.GetUserMe(getConinfo(c))
	if err != nil {
		if errors.Is(domain.ErrInvalidToken, err) {
			return forbidden(err, message("token is invalid."), needAuthorization(true))
		}
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, presentation.ConvdomainUserToUserRes(*user))
}

// HandleGetUsers ユーザーすべてを取得
func (h *Handlers) HandleGetUsers(c echo.Context) error {
	includeSuspend, _ := strconv.ParseBool(c.QueryParam("include-suspended"))

	users, err := h.Repo.GetAllUsers(includeSuspend, true, getConinfo(c))
	if err != nil {
		if errors.Is(domain.ErrInvalidToken, err) {
			return forbidden(err, message("token is invalid."), needAuthorization(true))
		}
		return judgeErrorResponse(err)
	}

	return c.JSON(http.StatusOK, presentation.ConvSPdomainUserToSUserRes(users))
}

func (h *Handlers) HandleGetiCal(c echo.Context) error {
	secret, err := h.Repo.GetMyiCalSecret(getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, struct {
		Secret string `json:"secret"`
	}{
		Secret: secret,
	})
}

func (h *Handlers) HandleUpdateiCal(c echo.Context) error {
	secret, err := h.Repo.ReNewMyiCalSecret(getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, struct {
		Secret string `json:"secret"`
	}{
		Secret: secret,
	})
}

// HandleSyncUser traQのユーザーとの同期をする
// 停止されているユーザーの`token`を削除して、
// 活動中のユーザーを追加する(userIDをDBに保存)
func (h *Handlers) HandleSyncUser(c echo.Context) error {
	repo, ok := h.Repo.(*production.Repository)
	if !ok {
		return internalServerError(errors.New("not implemented"))
	}
	err := repo.SyncUsers(getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.NoContent(http.StatusCreated)
}

// 権限のあるユーザーがないユーザに権限を付与
func (h *Handlers) HandleGrantPrivlege(c echo.Context) error {
	userID, err := getPathUserID(c)
	if err != nil {
		return notFound(err)
	}
	err = h.Repo.GrantPrivilege(userID)
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.NoContent(http.StatusCreated)
}
