package router

import (
	"fmt"
	"net/http"
	"room/middleware"
	repo "room/repository"
	"strconv"

	"github.com/labstack/echo"
)

// HandlePostReservation 部屋の使用宣言を作成
func HandlePostReservation(c echo.Context) error {
	rv := new(repo.Reservation)

	if err := c.Bind(&rv); err != nil {
		return err
	}

	rv.CreatedByRefer = middleware.GetRequestUser(c)
	if err := rv.AddCreatedBy(); err != nil {
		return err
	}

	// groupが存在するかチェックし依存関係を追加する
	if err := rv.Group.AddRelation(rv.GroupID); err != nil {
		return c.String(http.StatusBadRequest, "groupが存在しません"+fmt.Sprintln(rv.GroupID))
	}
	// roomが存在するかチェックし依存関係を追加する
	if err := rv.Room.AddRelation(rv.RoomID); err != nil {
		return c.String(http.StatusBadRequest, "roomが存在しません")
	}

	// dateを代入
	r := new(repo.Room)
	if err := db.First(&r, rv.RoomID).Error; err != nil {
		return err
	}
	// r.Date = 2018-08-10T00:00:00+09:00
	rv.Date = r.Date[:10]
	rv.Room.Date = rv.Room.Date[:10]

	err := rv.TimeConsistency()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := db.Create(&rv).Error; err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, rv)
}

// HandleGetReservations 部屋の使用宣言情報を取得
func HandleGetReservations(c echo.Context) error {
	reservations := []repo.Reservation{}

	values := c.QueryParams()

	reservations, err := repo.FindRvs(values)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "queryが正当でない")
	}

	return c.JSON(http.StatusOK, reservations)
}

// HandleDeleteReservation 部屋の使用宣言を削除
func HandleDeleteReservation(c echo.Context) error {
	rv := new(repo.Reservation)
	rv.ID, _ = strconv.Atoi(c.Param("reservationid"))

	traQID := middleware.GetRequestUser(c)
	belong, err := repo.CheckBelongToGroup(rv.ID, traQID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "reservationIDが正しくない")
	}
	if !belong {
		return c.String(http.StatusForbidden, "削除できるのは所属ユーザーのみです。")
	}

	if err := db.Delete(&rv).Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.NoContent(http.StatusOK)
}

// HandleUpdateReservation 部屋、開始時刻、終了時刻を更新
func HandleUpdateReservation(c echo.Context) error {
	rv := new(repo.Reservation)

	if err := c.Bind(&rv); err != nil {
		return err
	}
	rv.ID, _ = strconv.Atoi(c.Param("reservationid"))

	traQID := middleware.GetRequestUser(c)
	belong, err := repo.CheckBelongToGroup(rv.ID, traQID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "reservationIDが正しくない")
	}
	if !belong {
		return echo.NewHTTPError(http.StatusForbidden, "変更できるのは所属ユーザーのみです you are "+traQID)
	}

	// roomがあるか
	if err := rv.Room.AddRelation(rv.RoomID); err != nil {
		return c.String(http.StatusBadRequest, "roomが存在しません")
	}

	// r.Date = 2018-08-10T00:00:00+09:00
	rv.Room.Date = rv.Room.Date[:10]
	rv.Date = rv.Room.Date

	// roomid, timestart, timeendのみを変更(roomidに伴ってdateの変更する)
	if err := db.Model(&rv).Update(repo.Reservation{RoomID: rv.RoomID, Date: rv.Date, TimeStart: rv.TimeStart, TimeEnd: rv.TimeEnd}).Error; err != nil {
		fmt.Println("DB could not be updated")
		return err
	}

	if err := db.First(&rv, rv.ID).Error; err != nil {
		return err
	}

	if err := rv.Group.AddRelation(rv.GroupID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "GroupRelationを追加できませんでした")
	}

	if err := rv.AddCreatedBy(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "ReservationCreatedByを追加できませんでした")
	}

	return c.JSON(http.StatusOK, rv)
}
