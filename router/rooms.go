package router

import (
	"net/http"

	"github.com/traPtitech/knoQ/presentation"

	"github.com/labstack/echo/v4"
)

// HandlePostRoom traPで確保した部屋情報を作成
func (h *Handlers) HandlePostRoom(c echo.Context) error {
	var req presentation.RoomReq
	if err := c.Bind(&req); err != nil {
		return badRequest(err)
	}

	roomParams := presentation.ConvRoomReqTodomainWriteRoomParams(req)

	room, err := h.repo.CreateUnVerifiedRoom(roomParams, getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusCreated, presentation.ConvdomainRoomToRoomRes(*room))
}

// HandleCreateVerifedRooms csvを解析し、進捗部屋を作成
func (h *Handlers) HandleCreateVerifedRooms(c echo.Context) error {
	return c.NoContent(501)
}

// HandleGetRoom get one room
func (h *Handlers) HandleGetRoom(c echo.Context) error {
	roomID, err := getPathRoomID(c)
	if err != nil {
		return notFound(err)
	}

	room, err := h.repo.GetRoom(roomID)
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, presentation.ConvdomainRoomToRoomRes(*room))
}

// HandleGetRooms traPで確保した部屋情報を取得
func (h *Handlers) HandleGetRooms(c echo.Context) error {
	values := c.QueryParams()
	start, end, err := getTiemRange(values)
	if err != nil {
		return notFound(err)
	}
	rooms, err := h.repo.GetAllRooms(start, end)
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
	err = h.repo.DeleteRoom(roomID, getConinfo(c))
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

	err = h.repo.VerifyRoom(roomID, getConinfo(c))
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

	err = h.repo.UnVerifyRoom(roomID, getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.NoContent(http.StatusNoContent)
}
