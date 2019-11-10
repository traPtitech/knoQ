package router

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

// HandlePostRoom traPで確保した部屋情報を作成
func HandlePostRoom(c echo.Context) error {
	r := new(Room)
	if err := c.Bind(r); err != nil {
		return err
	}

	if err := db.Create(&r).Error; err != nil {
		return err
	}
	return c.JSON(http.StatusOK, r)
}

// HandleSetRooms Googleカレンダーから部屋情報を作成
func HandleSetRooms(c echo.Context) error {
	rooms, err := getEvents()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusCreated, rooms)
}

// HandleGetRooms traPで確保した部屋情報を取得
func HandleGetRooms(c echo.Context) error {
	r := []Room{}
	var err error
	id := c.QueryParam("id")
	begin := c.QueryParam("date_begin")
	end := c.QueryParam("date_end")

	if id == "" {
		r, err = findRoomsByTime(begin, end)
	} else {
		ID, _ := strconv.Atoi(id)
		err = db.First(&r, ID).Error
	}

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, r)
}

// HandleDeleteRoom traPで確保した部屋情報を削除
func HandleDeleteRoom(c echo.Context) error {
	r := new(Room)
	r.ID, _ = strconv.Atoi(c.Param("roomid"))

	if err := db.First(&r, r.ID).Error; err != nil {
		return c.String(http.StatusNotFound, "部屋が存在しない")
	}
	// 関連する予約を削除する
	if err := db.Where("room_id = ?", r.ID).Delete(&Reservation{}).Error; err != nil {
		fmt.Println(err)
	}

	if err := db.Delete(&r).Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.NoContent(http.StatusOK)
}
