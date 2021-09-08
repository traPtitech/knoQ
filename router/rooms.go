package router

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jszwec/csvutil"
	"github.com/traPtitech/knoQ/domain"
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

	room, err := h.Repo.CreateUnVerifiedRoom(roomParams, getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusCreated, presentation.ConvdomainRoomToRoomRes(*room))
}

// HandleCreateVerifedRooms csvを解析し、進捗部屋を作成
func (h *Handlers) HandleCreateVerifedRooms(c echo.Context) error {
	var req []presentation.RoomCSVReq
	userID, err := getRequestUserID(c)
	if err != nil {
		return notFound(err)
	}

	layout := "2006/01/02 15:04"

	buf := new(bytes.Buffer)
	io.Copy(buf, c.Request().Body)
	data := buf.Bytes()

	if err := csvutil.Unmarshal(data, &req); err != nil {
		return badRequest(err)
	}

	//構造体の変換

	for _, v := range req {
		var params domain.WriteRoomParams

		jst, _ := time.LoadLocation("Asia/Tokyo")
		params.Place = v.Location
		params.TimeStart, _ = time.ParseInLocation(layout, v.StartDate+""+v.StartTime, jst)
		params.TimeEnd, _ = time.ParseInLocation(layout, v.EndDate+""+v.EndTime, jst)
		params.Admins = []uuid.UUID{userID}

		_, err := h.Repo.CreateVerifiedRoom(params, getConinfo(c))

		if err != nil {
			return judgeErrorResponse(err)
		}

	}

	return c.NoContent(200)
}

// HandleGetRoom get one room
func (h *Handlers) HandleGetRoom(c echo.Context) error {
	roomID, err := getPathRoomID(c)
	if err != nil {
		return notFound(err)
	}

	room, err := h.Repo.GetRoom(roomID)
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
	rooms, err := h.Repo.GetAllRooms(start, end)
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
	err = h.Repo.DeleteRoom(roomID, getConinfo(c))
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

	err = h.Repo.VerifyRoom(roomID, getConinfo(c))
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

	err = h.Repo.UnVerifyRoom(roomID, getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.NoContent(http.StatusNoContent)
}
