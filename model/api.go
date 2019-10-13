package model

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

// HandleGetHello テスト用API
func HandleGetHello(c echo.Context) error {
	id := getRequestUser(c)                                      // リクエストしてきたユーザーのtraQID取得
	return c.String(http.StatusOK, fmt.Sprintf("hello %s!", id)) // レスポンスを返す
}

// HandleGetUserMe ヘッダー情報からuser情報を取得
func HandleGetUserMe(c echo.Context) error {
	traQID := getRequestUser(c)
	user, err := getUser(traQID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, user)
}

// HandleGetUsers ユーザーすべてを取得
func HandleGetUsers(c echo.Context) error {
	users := []User{}
	if err := db.Find(&users).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, users)
}

// RoomsAPI

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

// groupsAPI

// HandlePostGroup グループを作成
func HandlePostGroup(c echo.Context) error {
	g := new(Group)

	if err := c.Bind(&g); err != nil {
		return err
	}

	g.CreatedByRefer = getRequestUser(c)
	if err := g.AddCreatedBy(); err != nil {
		return err
	}

	// メンバーがdbにいるか
	if err := g.findMembers(); err != nil {
		return c.String(http.StatusBadRequest, "正しくないメンバーが含まれている")
	}

	if err := db.Create(&g).Error; err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprint(err))
	}

	return c.JSON(http.StatusCreated, g)
}

// HandleGetGroups グループを取得
func HandleGetGroups(c echo.Context) error {
	groups := []Group{}
	values := c.QueryParams()

	groups, err := findGroups(values)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, groups)
}

// HandleDeleteGroup グループを削除
func HandleDeleteGroup(c echo.Context) error {
	g := new(Group)
	g.ID, _ = strconv.Atoi(c.Param("groupid"))

	if err := db.First(&g, g.ID).Related(&g.Members, "Members").Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	// relationを削除
	if err := db.Model(&g).Association("Members").Clear().Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}
	// 予約情報を削除
	if err := db.Where("group_id = ?", g.ID).Delete(&Reservation{}).Error; err != nil {
		fmt.Println(err)
	}

	if err := db.Delete(&g).Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.NoContent(http.StatusOK)
}

// HandleUpdateGroup グループメンバー、グループ名を更新
func HandleUpdateGroup(c echo.Context) error {
	g := new(Group)

	if err := c.Bind(g); err != nil {
		return err
	}
	name := g.Name
	description := g.Description

	// メンバーがdbにいるか
	if err := g.findMembers(); err != nil {
		return c.String(http.StatusBadRequest, "正しくないメンバーが含まれている")
	}

	g.ID, _ = strconv.Atoi(c.Param("groupid"))
	if err := db.First(&g, g.ID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "アクセスしたgroupIDは存在しない")
	}
	// 作成者を取得
	if err := g.AddCreatedBy(); err != nil {
		return err
	}
	if getRequestUser(c) != g.CreatedByRefer {
		return echo.NewHTTPError(http.StatusForbidden, "作成者ではない")
	}

	// メンバーを置き換え
	if err := db.Model(&g).Association("Members").Replace(g.Members).Error; err != nil {
		return err
	}

	// グループ名を変更
	if err := db.Model(&g).Update("name", name).Error; err != nil {
		return err
	}
	fmt.Println(g.Name)
	// グループ詳細変更
	if err := db.Model(&g).Update("description", description).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, g)
}

// resrvations API

// HandlePostReservation 部屋の使用宣言を作成
func HandlePostReservation(c echo.Context) error {
	rv := new(Reservation)

	if err := c.Bind(&rv); err != nil {
		return err
	}

	rv.CreatedByRefer = getRequestUser(c)
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
	r := new(Room)
	if err := db.First(&r, rv.RoomID).Error; err != nil {
		return err
	}
	// r.Date = 2018-08-10T00:00:00+09:00
	rv.Date = r.Date[:10]
	rv.Room.Date = rv.Room.Date[:10]

	if err := db.Create(&rv).Error; err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, rv)
}

// HandleGetReservations 部屋の使用宣言情報を取得
func HandleGetReservations(c echo.Context) error {
	reservations := []Reservation{}

	values := c.QueryParams()

	reservations, err := findRvs(values)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "queryが正当でない")
	}

	return c.JSON(http.StatusOK, reservations)
}

// HandleDeleteReservation 部屋の使用宣言を削除
func HandleDeleteReservation(c echo.Context) error {
	rv := new(Reservation)
	rv.ID, _ = strconv.Atoi(c.Param("reservationid"))

	traQID := getRequestUser(c)
	belong, err := checkBelongToGroup(rv.ID, traQID)
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
	rv := new(Reservation)

	if err := c.Bind(&rv); err != nil {
		return err
	}
	rv.ID, _ = strconv.Atoi(c.Param("reservationid"))

	traQID := getRequestUser(c)
	belong, err := checkBelongToGroup(rv.ID, traQID)
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
	if err := db.Model(&rv).Update(Reservation{RoomID: rv.RoomID, Date: rv.Date, TimeStart: rv.TimeStart, TimeEnd: rv.TimeEnd}).Error; err != nil {
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
