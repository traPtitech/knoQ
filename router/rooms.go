package router

import (
	"fmt"
	"net/http"
	repo "room/repository"
	"strconv"

	"github.com/labstack/echo/v4"
)

// HandlePostRoom traPで確保した部屋情報を作成
func HandlePostRoom(c echo.Context) error {
	r := new(repo.Room)
	if err := c.Bind(r); err != nil {
		return err
	}

	if err := repo.DB.Create(&r).Error; err != nil {
		return err
	}
	return c.JSON(http.StatusOK, r)
}

// HandleSetRooms Googleカレンダーから部屋情報を作成
func HandleSetRooms(c echo.Context) error {
	rooms, err := repo.GetEvents()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusCreated, rooms)
}

// HandleGetRooms traPで確保した部屋情報を取得
func HandleGetRooms(c echo.Context) error {
	r := []repo.Room{}
	var err error
	id := c.QueryParam("id")
	begin := c.QueryParam("date_begin")
	end := c.QueryParam("date_end")

	if id == "" {
		r, err = repo.FindRoomsByTime(begin, end)
	} else {
		ID, _ := strconv.Atoi(id)
		err = repo.DB.First(&r, ID).Error
	}

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, r)
}

// HandleDeleteRoom traPで確保した部屋情報を削除
func HandleDeleteRoom(c echo.Context) error {
	r := new(repo.Room)
	r.ID, _ = strconv.Atoi(c.Param("roomid"))

	if err := repo.DB.First(&r, r.ID).Error; err != nil {
		return c.String(http.StatusNotFound, "部屋が存在しない")
	}
	// 関連する予約を削除する
	if err := repo.DB.Where("room_id = ?", r.ID).Delete(&repo.Event{}).Error; err != nil {
		fmt.Println(err)
	}

	if err := repo.DB.Delete(&r).Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.NoContent(http.StatusOK)
}
