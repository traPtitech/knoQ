package router

import (
	"net/http"
	repo "room/repository"

	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"

	"github.com/labstack/echo/v4"
)

// HandlePostEvent 部屋の使用宣言を作成
func (h *Handlers) HandlePostEvent(c echo.Context) error {
	req := new(EventReq)

	if err := c.Bind(&req); err != nil {
		return badRequest(message(err.Error()))
	}
	eventParams := new(repo.WriteEventParams)
	err := copier.Copy(&eventParams, req)
	if err != nil {
		return internalServerError()
	}

	eventParams.CreatedBy, _ = getRequestUserID(c)

	event, err := h.Repo.CreateEvent(*eventParams)
	if err != nil {
		return badRequest(message(err.Error()))
	}
	// add tag
	for _, reqTag := range req.Tags {
		tag, err := h.Repo.CreateOrGetTag(reqTag.Name)
		if err != nil {
			continue
		}
		tag.Locked = reqTag.Locked
		event.Tags = append(event.Tags, *tag)
		err = h.Repo.AddTagToEvent(event.ID, tag.ID, reqTag.Locked)
		if err != nil {
			return internalServerError()
		}
	}

	return c.JSON(http.StatusCreated, formatEventRes(event))
}

// HandleGetEvent get one event
func (h *Handlers) HandleGetEvent(c echo.Context) error {
	eventID, err := getRequestEventID(c)
	if err != nil {
		return notFound()
	}

	event, err := h.Repo.GetEvent(eventID)
	if err != nil {
		return internalServerError()
	}
	res := formatEventRes(event)
	return c.JSON(http.StatusOK, res)
}

// HandleGetEvents 部屋の使用宣言情報を取得
func (h *Handlers) HandleGetEvents(c echo.Context) error {
	values := c.QueryParams()

	start, end, err := getTiemRange(values)
	if err != nil {
		return notFound()
	}

	events, err := h.Repo.GetAllEvents(&start, &end)
	if err != nil {
		return internalServerError()
	}
	res := formatEventsRes(events)
	return c.JSON(http.StatusOK, res)
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

// HandleUpdateEvent 任意の要素を変更
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
	res := formatEventRes(event)
	return c.JSON(http.StatusOK, res)
}

func (h *Handlers) HandleAddEventTag(c echo.Context) error {
	var req TagRelationReq
	if err := c.Bind(&req); err != nil {
		return badRequest()
	}
	eventID, err := getRequestEventID(c)
	if err != nil {
		return notFound(message(err.Error()))
	}
	tag, err := h.Repo.CreateOrGetTag(req.Name)
	if err != nil {
		return internalServerError()
	}
	err = h.Repo.AddTagToEvent(eventID, tag.ID, false)
	if err != nil {
		return internalServerError()
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandleDeleteEventTag(c echo.Context) error {
	eventID, err := getRequestEventID(c)
	if err != nil {
		return notFound(message(err.Error()))
	}
	tagName := c.Param("tagName")
	tag, err := h.Repo.CreateOrGetTag(tagName)
	if err != nil {
		return internalServerError()
	}
	err = h.Repo.DeleteTagInEvent(eventID, tag.ID, false)
	if err != nil {
		return internalServerError()
	}

	return c.NoContent(http.StatusNoContent)
}
