package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

// GetHello テスト用API
func GetHello(c echo.Context) error {
	id := getRequestUser(c)                                      // リクエストしてきたユーザーのtraQID取得
	return c.String(http.StatusOK, fmt.Sprintf("hello %s!", id)) // レスポンスを返す
}

func SaveRoom(c echo.Context) error {
	r := new(Room)
	if err := c.Bind(r); err != nil {
		return err
	}

	if err := db.Create(&r).Error; err != nil {
		return err
	}
	return c.JSON(http.StatusOK, r)
}

func GetRooms(c echo.Context) error {
	r := []Room{}
	if err := db.Find(&r).Error; err != nil {
		return err
	}
	return c.JSON(http.StatusOK, r)
}

func DeleteRoom(c echo.Context) error {
	r := new(Room)
	r.ID, _ = strconv.Atoi(c.Param("roomid"))

	if err := db.First(&r, r.ID).Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	err := db.Delete(&r)
	if err.Error != nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.NoContent(http.StatusOK)
}
