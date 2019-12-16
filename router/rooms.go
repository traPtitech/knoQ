package router

import (
	"fmt"
	"net/http"
	repo "room/repository"
	"strconv"

	"github.com/jinzhu/gorm"

	"github.com/labstack/echo/v4"
)

// HandlePostRoom traPで確保した部屋情報を作成
func HandlePostRoom(c echo.Context) error {
	r := new(repo.Room)
	if err := c.Bind(r); err != nil {
		return badRequest()
	}

	if err := repo.DB.Create(&r).Error; err != nil {
		return internalServerError()
	}
	return c.JSON(http.StatusOK, r)
}

// HandleSetRooms Googleカレンダーから部屋情報を作成
func HandleSetRooms(c echo.Context) error {
	rooms, err := repo.GetEvents()
	if err != nil {
		return internalServerError()
	}
	return c.JSON(http.StatusCreated, rooms)
}

// HandleGetRoom get one room
func HandleGetRoom(c echo.Context) error {
	r := new(repo.Room)
	r.ID, _ = strconv.ParseUint(c.Param("roomid"), 10, 64)

	if err := r.Read(); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return notFound()
		}
		return internalServerError()
	}
	return c.JSON(http.StatusOK, r)
}

// HandleGetRooms traPで確保した部屋情報を取得
func HandleGetRooms(c echo.Context) error {
	rooms := []repo.Room{}

	values := c.QueryParams()
	rooms, err := repo.FindRooms(values)
	if err != nil {
		return internalServerError()
	}

	return c.JSON(http.StatusOK, rooms)
}

// HandleDeleteRoom traPで確保した部屋情報を削除
func HandleDeleteRoom(c echo.Context) error {
	r := new(repo.Room)
	r.ID, _ = strconv.ParseUint(c.Param("roomid"), 10, 64)

	if err := repo.DB.First(&r, r.ID).Error; err != nil {
		return notFound(message(fmt.Sprintf("RoomID: %v does not exist.", r.ID)))
	}

	if err := repo.DB.Delete(&r).Error; err != nil {
		return internalServerError()
	}

	return c.NoContent(http.StatusOK)
}
