package router

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/router/presentation"

	"github.com/labstack/echo/v4"
)

// HandleGetUserMe ヘッダー情報からuser情報を取得
// 認証状態を確認
func (h *Handlers) HandleGetUserMe(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	user, err := h.Service.GetUserMe(ctx, reqID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) {
			return forbidden(err, message("token is invalid."), needAuthorization(true))
		}
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, presentation.ConvdomainUserToUserRes(*user))
}

// HandleGetUsers ユーザーすべてを取得
func (h *Handlers) HandleGetUsers(c echo.Context) error {
	includeSuspend, _ := strconv.ParseBool(c.QueryParam("include-suspended"))
	ctx := c.Request().Context()

	users, err := h.Service.GetAllUsers(ctx, includeSuspend, true)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) {
			return forbidden(err, message("token is invalid."), needAuthorization(true))
		}
		return judgeErrorResponse(err)
	}

	return c.JSON(http.StatusOK, presentation.ConvSPdomainUserToSPUserRes(users))
}

func (h *Handlers) HandleGetiCal(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	secret, err := h.Service.GetMyiCalSecret(ctx, reqID)
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
	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	secret, err := h.Service.ReNewMyiCalSecret(ctx, reqID)
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
	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	err := h.Service.SyncUsers(ctx, reqID)
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.NoContent(http.StatusCreated)
}

// 権限のあるユーザーがないユーザーに権限を付与
func (h *Handlers) HandleGrantPrivilege(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getPathUserID(c)
	if err != nil {
		return notFound(err)
	}
	err = h.Service.GrantPrivilege(ctx, userID)
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.NoContent(http.StatusCreated)
}
