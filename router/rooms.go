package router

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/router/presentation"

	"github.com/labstack/echo/v4"
)

// HandlePostRoom traPで確保した部屋情報を作成
func (h *Handlers) HandlePostRoom(c echo.Context) error {
	var req presentation.RoomReq
	if err := c.Bind(&req); err != nil {
		return badRequest(err)
	}

	roomParams := presentation.ConvRoomReqTodomainWriteRoomParams(req)
	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	room, err := h.Service.CreateUnVerifiedRoom(ctx, reqID, roomParams)
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusCreated, presentation.ConvdomainRoomToRoomRes(*room))
}

// HandleCreateVerifedRooms csvを解析し、進捗部屋を作成
func (h *Handlers) HandleCreateVerifedRooms(c echo.Context) error {
	userID, err := getRequestUserID(c)
	if err != nil {
		return notFound(err)
	}

	var req []presentation.RoomCSVReq
	if err := c.Bind(&req); err != nil {
		return badRequest(err)
	}

	// 構造体の変換
	RoomsRes := make([]presentation.RoomRes, 0, len(req))
	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	for _, v := range req {

		params, err := presentation.ChangeRoomCSVReqTodomainWriteRoomParams(v, userID)
		if err != nil {
			return badRequest(err)
		}

		room, err := h.Service.CreateVerifiedRoom(ctx, reqID, *params)
		if err != nil {
			return judgeErrorResponse(err)
		}

		RoomsRes = append(RoomsRes, presentation.ConvdomainRoomToRoomRes(*room))

	}
	return c.JSON(http.StatusCreated, RoomsRes)
}

// HandleGetRoom get one room
func (h *Handlers) HandleGetRoom(c echo.Context) error {
	roomID, err := getPathRoomID(c)
	if err != nil {
		return notFound(err)
	}
	values := c.QueryParams()
	excludeEventID, err := presentation.GetExcludeEventID(values)
	if err != nil {
		return badRequest(err)
	}

	ctx := c.Request().Context()
	room, err := h.Service.GetRoom(ctx, roomID, excludeEventID)
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, presentation.ConvdomainRoomToRoomRes(*room))
}

// HandleGetRooms traPで確保した部屋情報を取得
func (h *Handlers) HandleGetRooms(c echo.Context) error {
	values := c.QueryParams()
	start, end, err := presentation.GetTiemRange(values)
	if err != nil {
		return notFound(err)
	}
	excludeEventID, err := presentation.GetExcludeEventID(values)
	if err != nil {
		return badRequest(err)
	}

	ctx := c.Request().Context()
	rooms, err := h.Service.GetAllRooms(ctx, start, end, excludeEventID)
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, presentation.ConvSPdomainRoomToSPRoomRes(rooms))
}

// HandleDeleteRoom traPで確保した部屋情報を削除
func (h *Handlers) HandleDeleteRoom(c echo.Context) error {
	roomID, err := getPathRoomID(c)
	if err != nil {
		return notFound(err)
	}

	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	err = h.Service.DeleteRoom(ctx, reqID, roomID)
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandleVerifyRoom(c echo.Context) error {
	roomID, err := getPathRoomID(c)
	if err != nil {
		return notFound(err)
	}

	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	err = h.Service.VerifyRoom(ctx, reqID, roomID)
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandleUnVerifyRoom(c echo.Context) error {
	roomID, err := getPathRoomID(c)
	if err != nil {
		return notFound(err)
	}

	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	err = h.Service.UnVerifyRoom(ctx, reqID, roomID)
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.NoContent(http.StatusNoContent)
}
