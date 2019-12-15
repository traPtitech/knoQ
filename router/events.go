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
	event := new(repo.Event)

	if err := c.Bind(event); err != nil {
		return badRequest()
	}
	event.Group.ID = event.GroupID
	event.Room.ID = event.RoomID

	event.CreatedBy = getRequestUser(c).TRAQID

	err := event.Create()
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.JSON(http.StatusCreated, event)
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

	events, err := repo.FindEvents(values)
	if err != nil {
		return internalServerError()
	}

	return c.JSON(http.StatusOK, events)
}

// HandleDeleteEvent 部屋の使用宣言を削除
func HandleDeleteEvent(c echo.Context) error {
	event := new(repo.Event)
	var err error
	event.ID, err = getRequestEventID(c)

	if err = event.Delete(); err != nil {
		return internalServerError()
	}
	return c.NoContent(http.StatusOK)
}

// HandleUpdateEvent 部屋、開始時刻、終了時刻を更新
func HandleUpdateEvent(c echo.Context) error {
	event := new(repo.Event)

	if err := c.Bind(event); err != nil {
		return badRequest(message(err.Error()))
	}

	var err error
	event.ID, err = getRequestEventID(c)
	if err != nil {
		return internalServerError()
	}

	err = event.Update()
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return badRequest()
		}
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
	if err := repo.MatchTag(tag, "event"); err != nil {
		internalServerError()
	}
	var err error
	event.ID, err = getRequestEventID(c)
	if err != nil {
		internalServerError()
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
	eventTag := new(repo.EventTag)
	event := new(repo.Event)
	var err error
	event.ID, err = getRequestEventID(c)
	if err != nil {
		internalServerError()
	}
	eventTag.TagID, err = strconv.ParseUint(c.Param("tagid"), 10, 64)
	if err != nil || eventTag.TagID == 0 {
		return notFound(message(fmt.Sprintf("TagID: %v does not exist.", c.Param("tagid"))))
	}
	eventTag.EventID = event.ID

	if err := repo.DB.Debug().First(&eventTag).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return notFound(message(fmt.Sprintf("This event does not have TagID: %v.", eventTag.TagID)))
		}
		internalServerError()
	}
	if eventTag.Locked {
		return forbidden(message("This tag is locked."), specification("This api can delete non-locked tags"))
	}

	if err := repo.DB.Debug().Where("locked = ?", false).Delete(&repo.EventTag{EventID: event.ID, TagID: eventTag.TagID}).Error; err != nil {
		internalServerError()
	}
	if err := event.Read(); err != nil {
		internalServerError()
	}
	return c.JSON(http.StatusOK, event)

}
