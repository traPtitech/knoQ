package router

import (
	"fmt"
	"net/http"
	repo "room/repository"
	"strconv"

	"github.com/jinzhu/gorm"

	"github.com/labstack/echo/v4"
)

// HandlePostEvent 部屋の使用宣言を作成
func HandlePostEvent(c echo.Context) error {
	rv := new(repo.Event)

	if err := c.Bind(rv); err != nil {
		return badRequest()
	}

	rv.CreatedBy = getRequestUser(c).TRAQID

	// groupが存在するかチェックし依存関係を追加する
	if err := rv.Group.AddRelation(rv.GroupID); err != nil {
		return badRequest(message(fmt.Sprintf("GroupID: %v does not exist.", rv.GroupID)))
	}
	// roomが存在するかチェックし依存関係を追加する
	if err := rv.Room.AddRelation(rv.RoomID); err != nil {
		return badRequest(message(fmt.Sprintf("RoomID: %v does not exist.", rv.RoomID)))
	}

	// format
	rv.Room.Date = rv.Room.Date[:10]

	err := rv.TimeConsistency()
	if err != nil {
		return badRequest(message(err.Error()))
	}

	err = repo.MatchEventTags(rv.Tags)
	if err != nil {
		internalServerError()
	}

	err = rv.Create()
	if err != nil {
		internalServerError()
	}

	return c.JSON(http.StatusCreated, rv)
}

// HandleGetEvent get one event
func HandleGetEvent(c echo.Context) error {
	event := new(repo.Event)
	var err error
	event.ID, err = getRequestEventID(c)
	if err != nil {
		return internalServerError()
	}

	if err := event.Read(); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return notFound()
		}
		return internalServerError()
	}
	return c.JSON(http.StatusOK, event)
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
	e := new(repo.Event)
	var err error
	e.ID, err = strconv.ParseUint(c.Param("eventid"), 10, 64)
	if err != nil || e.ID == 0 {
		return notFound(message(fmt.Sprintf("EventID: %v does not exist.", c.Param("eventid"))))
	}
	if err = e.Delete(); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return notFound(message(fmt.Sprintf("EventID: %v does not exist.", e.ID)))
		}
		return internalServerError()
	}
	return c.NoContent(http.StatusOK)
}

// HandleUpdateEvent 部屋、開始時刻、終了時刻を更新
func HandleUpdateEvent(c echo.Context) error {
	event := new(repo.Event)
	nowEvent := new(repo.Event)
	if err := nowEvent.Read(); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			notFound()
		}
		return internalServerError()
	}

	if err := c.Bind(event); err != nil {
		return badRequest(message(err.Error()))
	}

	var err error
	event.ID, err = strconv.ParseUint(c.Param("eventid"), 10, 64)
	if err != nil || event.ID == 0 {
		return notFound(message(fmt.Sprintf("EventID: %v does not exist.", c.Param("eventid"))))
	}

	// groupが存在するかチェックし依存関係を追加する
	if err := event.Group.AddRelation(event.GroupID); err != nil {
		return badRequest(message(fmt.Sprintf("GroupID: %v does not exist.", event.GroupID)))
	}
	// roomが存在するかチェックし依存関係を追加する
	if err := event.Room.AddRelation(event.RoomID); err != nil {
		return badRequest(message(fmt.Sprintf("RoomID: %v does not exist.", event.RoomID)))
	}

	// r.Date = 2018-08-10T00:00:00+09:00
	event.Room.Date = event.Room.Date[:10]
	err = event.TimeConsistency()
	if err != nil {
		return badRequest(message(err.Error()))
	}

	event.CreatedBy = nowEvent.CreatedBy
	event.CreatedAt = nowEvent.CreatedAt

	if err := repo.DB.Debug().Save(&event).Error; err != nil {
		return internalServerError()
	}

	if err := event.Read(); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return notFound()
		}
		return internalServerError()
	}

	return c.JSON(http.StatusOK, event)
}

func HandleAddEventTag(c echo.Context) error {
	tag := new(repo.Tag)
	event := new(repo.Event)
	if err := c.Bind(tag); err != nil {
		badRequest()
	}
	if err := repo.MatchEventTag(tag); err != nil {
		internalServerError()
	}
	var err error
	event.ID, err = strconv.ParseUint(c.Param("eventid"), 10, 64)
	if err != nil || event.ID == 0 {
		return notFound(message(fmt.Sprintf("EventID: %v does not exist.", c.Param("eventid"))))
	}

	if err := repo.DB.Create(&repo.EventTag{EventID: event.ID, TagID: tag.ID}).Error; err != nil {
		internalServerError()
	}

	if err := event.Read(); err != nil {
		internalServerError()
	}
	return c.JSON(http.StatusOK, event)
}

func HandleDeleteEventTag(c echo.Context) error {
	event := new(repo.Event)

	eventID, err := strconv.ParseUint(c.Param("eventid"), 10, 64)
	if err != nil || eventID == 0 {
		return notFound(message(fmt.Sprintf("EventID: %v does not exist.", c.Param("eventid"))))
	}
	tagID, err := strconv.ParseUint(c.Param("tagid"), 10, 64)
	if err != nil || tagID == 0 {
		return notFound(message(fmt.Sprintf("TagID: %v does not exist.", c.Param("tagid"))))
	}

	if err := repo.DB.Delete(&repo.EventTag{EventID: event.ID, TagID: tagID}).Error; err != nil {
		internalServerError()
	}
	if err := event.Read(); err != nil {
		internalServerError()
	}
	return c.JSON(http.StatusOK, event)

}
