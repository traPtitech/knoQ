package router

import (
	"fmt"
	"net/http"
	repo "room/repository"
	"strconv"

	"github.com/labstack/echo/v4"
)

// HandlePostEvent 部屋の使用宣言を作成
func HandlePostEvent(c echo.Context) error {
	rv := new(repo.Event)

	if err := c.Bind(&rv); err != nil {
		return err
	}

	rv.CreatedByRefer = getRequestUser(c).TRAQID
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
	if err := repo.DB.First(&r, rv.RoomID).Error; err != nil {
		return err
	}
	// r.Date = 2018-08-10T00:00:00+09:00
	rv.Date = r.Date[:10]
	rv.Room.Date = rv.Room.Date[:10]

	err := rv.TimeConsistency()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := repo.DB.Create(&rv).Error; err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, rv)
}

// HandleGetEvents 部屋の使用宣言情報を取得
func HandleGetEvents(c echo.Context) error {
	events := []repo.Event{}

	values := c.QueryParams()

	events, err := repo.FindRvs(values)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "queryが正当でない")
	}

	return c.JSON(http.StatusOK, events)
}

// HandleDeleteEvent 部屋の使用宣言を削除
func HandleDeleteEvent(c echo.Context) error {
	rv := new(repo.Event)
	rv.ID, _ = strconv.Atoi(c.Param("eventid"))

	if err := repo.DB.Delete(&rv).Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.NoContent(http.StatusOK)
}

// HandleUpdateEvent 部屋、開始時刻、終了時刻を更新
func HandleUpdateEvent(c echo.Context) error {
	rv := new(repo.Event)

	if err := c.Bind(&rv); err != nil {
		return err
	}
	rv.ID, _ = strconv.Atoi(c.Param("eventid"))

	// roomがあるか
	if err := rv.Room.AddRelation(rv.RoomID); err != nil {
		return c.String(http.StatusBadRequest, "roomが存在しません")
	}

	// r.Date = 2018-08-10T00:00:00+09:00
	rv.Room.Date = rv.Room.Date[:10]
	rv.Date = rv.Room.Date

	// roomid, timestart, timeendのみを変更(roomidに伴ってdateの変更する)
	if err := repo.DB.Model(&rv).Update(repo.Event{RoomID: rv.RoomID, Date: rv.Date, TimeStart: rv.TimeStart, TimeEnd: rv.TimeEnd}).Error; err != nil {
		fmt.Println("DB could not be updated")
		return err
	}

	if err := repo.DB.First(&rv, rv.ID).Error; err != nil {
		return err
	}

	if err := rv.Group.AddRelation(rv.GroupID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "GroupRelationを追加できませんでした")
	}

	if err := rv.AddCreatedBy(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "EventCreatedByを追加できませんでした")
	}

	return c.JSON(http.StatusOK, rv)
}
