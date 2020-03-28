package router

import (
	"net/http"
	repo "room/repository"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/copier"

	"github.com/labstack/echo/v4"
)

// HandlePostRoom traPで確保した部屋情報を作成
func (h *Handlers) HandlePostRoom(c echo.Context) error {
	req := new(RoomReq)
	if err := c.Bind(&req); err != nil {
		return badRequest()
	}
	roomParams := new(repo.WriteRoomParams)
	err := copier.Copy(&roomParams, req)
	if err != nil {
		return internalServerError()
	}
	roomParams.Public = true

	room, err := h.Repo.CreateRoom(*roomParams)
	if err != nil {
		return internalServerError()
	}
	return c.JSON(http.StatusOK, formatRoomRes(room))
}

// HandleSetRooms Googleカレンダーから部屋情報を作成
func (h *Handlers) HandleSetRooms(c echo.Context) error {
	now := time.Now()
	googleRooms, err := h.ExternalRoomRepo.GetAllRooms(&now, nil)
	if err != nil {
		return internalServerError()
	}
	res := make([]*RoomRes, 0)
	for _, room := range googleRooms {
		roomParams := new(repo.WriteRoomParams)
		err := copier.Copy(&roomParams, room)
		if err != nil {
			return internalServerError()
		}

		room, err := h.Repo.CreateRoom(*roomParams)
		if err != nil {
			return internalServerError()
		}
		res = append(res, formatRoomRes(room))
	}

	return c.JSON(http.StatusCreated, res)
}

// HandleGetRoom get one room
func (h *Handlers) HandleGetRoom(c echo.Context) error {
	roomID, err := uuid.FromString(c.Param("roomid"))
	if err != nil {
		return notFound()
	}

	room, err := h.Repo.GetRoom(roomID)
	if err != nil {
		return notFound()
	}
	return c.JSON(http.StatusOK, formatRoomRes(room))
}

// HandleGetRooms traPで確保した部屋情報を取得
func (h *Handlers) HandleGetRooms(c echo.Context) error {
	values := c.QueryParams()
	start, end, err := getTiemRange(values)
	if err != nil {
		return notFound()
	}
	rooms, err := h.Repo.GetAllRooms(&start, &end)
	if err != nil {
		return internalServerError()
	}
	res := make([]*RoomRes, len(rooms))
	for i, r := range rooms {
		res[i] = formatRoomRes(r)
	}
	return c.JSON(http.StatusCreated, res)

}

// HandleDeleteRoom traPで確保した部屋情報を削除
func (h *Handlers) HandleDeleteRoom(c echo.Context) error {
	roomID, err := uuid.FromString(c.Param("roomid"))
	if err != nil {
		return notFound()
	}
	err = h.Repo.DeleteRoom(roomID, true)
	if err != nil {
		return notFound()
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandlePostPrivateRoom(c echo.Context) error {
	req := new(RoomReq)
	if err := c.Bind(&req); err != nil {
		return badRequest()
	}
	roomParams := new(repo.WriteRoomParams)
	err := copier.Copy(&roomParams, req)
	if err != nil {
		return internalServerError()
	}

	roomParams.Public = false

	room, err := h.Repo.CreateRoom(*roomParams)
	if err != nil {
		return internalServerError()
	}
	return c.JSON(http.StatusOK, formatRoomRes(room))
}

func (h *Handlers) HandleDeletePrivateRoom(c echo.Context) error {
	roomID, err := uuid.FromString(c.Param("roomid"))
	if err != nil {
		return notFound()
	}
	err = h.Repo.DeleteRoom(roomID, false)
	if err != nil {
		return notFound()
	}

	return c.NoContent(http.StatusNoContent)

}
