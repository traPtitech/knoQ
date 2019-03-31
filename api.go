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

// RoomsAPI

// PostRoom traPで確保した部屋情報を作成
func PostRoom(c echo.Context) error {
	r := new(Room)
	if err := c.Bind(r); err != nil {
		return err
	}

	if err := db.Create(&r).Error; err != nil {
		return err
	}
	return c.JSON(http.StatusOK, r)
}

// GetRooms traPで確保した部屋情報を取得
func GetRooms(c echo.Context) error {
	r := []Room{}
	begin := c.QueryParam("date_begin")
	end := c.QueryParam("date_end")

	if begin == "" && end == "" {
		if err := db.Find(&r).Error; err != nil {
			return err
		}
	} else if end == "" {
		if err := db.Where("date >= ?", begin).Find(&r).Error; err != nil {
			return err
		}
	} else if begin == "" {
		if err := db.Where("date <= ?", end).Find(&r).Error; err != nil {
			return err
		}
	} else {
		if err := db.Where("date BETWEEN ? AND ?", begin, end).Find(&r).Error; err != nil {
			return err
		}
	}

	return c.JSON(http.StatusOK, r)
}

// DeleteRoom traPで確保した部屋情報を削除
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

// PostGroup グループを作成
func PostGroup(c echo.Context) error {
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

// GetGroups グループを取得
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

// DeleteGroup グループを削除
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

// UpdateGroup グループメンバーを更新
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

// resrvations API

// PostReservation 部屋の使用宣言を作成
func PostReservation(c echo.Context) error {
	rv := new(Reservation)

	if err := c.Bind(&rv); err != nil {
		return err
	}

	// groupがあるか
	if err := checkGroup(rv.GroupID); err != nil {
		return c.String(http.StatusBadRequest, "groupが存在しません"+fmt.Sprintln(rv.GroupID))
	}
	// roomがあるか
	if err := checkRoom(rv.RoomID); err != nil {
		return c.String(http.StatusBadRequest, "roomが存在しません")
	}

	// dateを代入
	r := new(Room)
	if err := db.First(&r, rv.RoomID).Error; err != nil {
		return err
	}
	// r.Date = 2018-08-10T00:00:00+09:00
	rv.Date = r.Date[:10]

	if err := db.Create(&rv).Error; err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, rv)
}

// GetReservations 部屋の使用宣言情報を取得
func GetReservations(c echo.Context) error {
	reservations := []Reservation{}
	cmd := db
	rv := new(Reservation)
	/*
		traqID := c.QueryParam("userid")
		if traqID != "" {
			cmd = cmd.Where("traq_id = ?", traqID)
		}
	*/

	if c.QueryParam("groupid") != "" {
		rv.GroupID, _ = strconv.Atoi(c.QueryParam("groupid"))
		cmd = cmd.Where("group_id = ?", rv.GroupID)
	}
	begin := c.QueryParam("date_begin")
	if begin != "" {
		cmd = cmd.Where("date >= ?", begin)
	}
	end := c.QueryParam("date_end")
	if end != "" {
		cmd = cmd.Where("date <= ?", end)
	}

	if err := cmd.Find(&reservations).Error; err != nil {
		return err
	}
	return c.JSON(http.StatusOK, reservations)
}

// DeleteReservation 部屋の使用宣言を削除
func DeleteReservation(c echo.Context) error {
	rv := new(Reservation)
	rv.ID, _ = strconv.Atoi(c.Param("reservationid"))

	if err := db.Delete(&rv).Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.NoContent(http.StatusOK)
}

// UpdateReservation 部屋、開始時刻、終了時刻を更新
func UpdateReservation(c echo.Context) error {
	rv := new(Reservation)

	if err := c.Bind(&rv); err != nil {
		return err
	}
	rv.ID, _ = strconv.Atoi(c.Param("reservationid"))

	// roomがあるか
	if err := checkRoom(rv.RoomID); err != nil {
		return c.String(http.StatusBadRequest, "roomが存在しません")
	}

	// dateを代入
	r := new(Room)
	if err := db.First(&r, rv.RoomID).Error; err != nil {
		return err
	}
	// r.Date = 2018-08-10T00:00:00+09:00
	rv.Date = r.Date[:10]

	// roomid, timestart, timeendのみを変更
	if err := db.Model(&rv).Update(Reservation{RoomID: rv.RoomID, TimeStart: rv.TimeStart, TimeEnd: rv.TimeEnd}).Error; err != nil{
		return err
	}

	if err := db.First(&rv, rv.ID).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, rv)
}
