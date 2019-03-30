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
	begin := c.QueryParam("date_begin")
	end := c.QueryParam("date_end")

	if begin == "" && end == "" {
		if err := db.Find(&r).Error; err != nil {
			return err
		}
	}else if end == ""{
		if err := db.Where("date >= ?", begin).Find(&r).Error; err != nil {
			return err
		} 
	}else if begin == ""{
		if err := db.Where("date <= ?", end).Find(&r).Error; err != nil {
			return err
		} 	
	}else {
		if err := db.Where("date BETWEEN ? AND ?", begin, end).Find(&r).Error; err != nil {
			return err
		}
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

// groupsAPI

func SaveGroup(c echo.Context) error {
	g := new(Group)

	if err := c.Bind(&g); err != nil {
		return err
	}

	// メンバーがdbにいるか
	if err := checkMembers(g); err != nil {
		return c.String(http.StatusBadRequest, "正しくないメンバーが含まれている")
	}

	if err := db.Create(&g).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, g)
}

func GetGroups(c echo.Context) error {
	groups := []Group{}
	traqID := c.QueryParam("userid")

	if err := db.Find(&groups).Error; err != nil {
		return err
	}

	resGroups := []Group{}
	for _, g := range groups {
		if err := db.First(&g, g.ID).Related(&g.Members, "Members").Error; err != nil {
			return err
		}

		for _, user := range g.Members {
			if user.TRAQID == traqID || traqID == "" {
				resGroups = append(resGroups, g)
				break
			}
		}
	}
	return c.JSON(http.StatusOK, resGroups)
}

func DeleteGroup(c echo.Context) error {
	g := new(Group)
	g.ID, _ = strconv.Atoi(c.Param("groupid"))

	if err := db.First(&g, g.ID).Related(&g.Members, "Members").Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	if err := db.Model(&g).Association("Members").Clear().Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	if err := db.Delete(&g).Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.NoContent(http.StatusOK)
}

func UpdateGroup(c echo.Context) error {
	g := new(Group)

	if err := c.Bind(g); err != nil {
		return err
	}

	// メンバーがdbにいるか
	if err := checkMembers(g); err != nil {
		return c.String(http.StatusBadRequest, "正しくないメンバーが含まれている")
	}

	g.ID, _ = strconv.Atoi(c.Param("groupid"))

	// メンバーを変更
	if err := db.Model(&g).Association("Members").Replace(g.Members).Error; err != nil {
		return err
	}

	if err := db.Save(&g).Error; err != nil {
		return err
	}

	if err := db.First(&g, g.ID).Error; err != nil {
		return err
	}
	return c.JSON(http.StatusOK, g)
}