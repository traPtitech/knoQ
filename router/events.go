package router

import (
	"net/http"
	repo "room/repository"
	"room/router/service"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/copier"

	"github.com/labstack/echo/v4"
)

// HandlePostEvent 部屋の使用宣言を作成
func (h *Handlers) HandlePostEvent(c echo.Context) error {
	var req service.EventReq
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

	return c.JSON(http.StatusCreated, service.FormatEventRes(event))
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
	res := service.FormatEventRes(event)
	return c.JSON(http.StatusOK, res)
}

// HandleGetEvents 部屋の使用宣言情報を取得
func (h *Handlers) HandleGetEvents(c echo.Context) error {
	values := c.QueryParams()

	start, end, err := getTiemRange(values)
	if err != nil {
		return badRequest(message("invalid time"))
	}

	events, err := h.Repo.GetAllEvents(&start, &end)
	if err != nil {
		return internalServerError()
	}
	res := service.FormatEventsRes(events)
	return c.JSON(http.StatusOK, res)
}

// HandleGetEvents groupidの仕様宣言を取得
func (h *Handlers) HandleGetEventsByGroupID(c echo.Context) error {
	groupID, err := getRequestGroupID(c)
	if err != nil {
		return notFound()
	}
	events, err := h.Repo.GetEventsByGroupIDs([]uuid.UUID{groupID})
	if err != nil {
		return internalServerError()
	}
	return c.JSON(http.StatusOK, service.FormatEventsRes(events))

}

// HandleDeleteEvent 部屋の使用宣言を削除
func (h *Handlers) HandleDeleteEvent(c echo.Context) error {
	eventID, err := getRequestEventID(c)
	if err != nil {
		return notFound()
	}

	if err = h.Repo.DeleteEvent(eventID); err != nil {
		return internalServerError()
	}
	return c.NoContent(http.StatusNoContent)
}

// HandleUpdateEvent 任意の要素を変更
func (h *Handlers) HandleUpdateEvent(c echo.Context) error {
	var req service.EventReq
	if err := c.Bind(&req); err != nil {
		return badRequest(message(err.Error()))
	}
	eventParams := new(repo.WriteEventParams)
	err := copier.Copy(&eventParams, req)
	if err != nil {
		return internalServerError()
	}
	eventID, err := getRequestEventID(c)
	if err != nil {
		return notFound()
	}
	eventParams.CreatedBy, _ = getRequestUserID(c)

	event, err := h.Repo.UpdateEvent(eventID, *eventParams)
	if err != nil {
		return internalServerError()
	}
	// update tag
	tagsParams := make([]repo.WriteTagRelationParams, 0)
	for _, reqTag := range req.Tags {
		tag, err := h.Repo.CreateOrGetTag(reqTag.Name)
		if err != nil {
			continue
		}
		tagsParams = append(tagsParams, repo.WriteTagRelationParams{
			ID:     tag.ID,
			Locked: reqTag.Locked,
		})
		event.Tags = append(event.Tags, *tag)
	}
	err = h.Repo.UpdateTagsInEvent(eventID, tagsParams)
	if err != nil {
		return err
	}

	res := service.FormatEventRes(event)
	return c.JSON(http.StatusOK, res)
}

func (h *Handlers) HandleAddEventTag(c echo.Context) error {
	var req service.TagRelationReq
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
	tag, err := h.Repo.GetTagByName(tagName)
	if err != nil {
		return internalServerError()
	}
	err = h.Repo.DeleteTagInEvent(eventID, tag.ID, false)
	if err != nil {
		return internalServerError()
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandleGetMeEvents(c echo.Context) error {
	userID, _ := getRequestUserID(c)

	token, _ := getRequestUserToken(c)
	groupIDs, err := h.Dao.GetUserBelongingGroupIDs(token, userID)
	if err != nil {
		return internalServerError()
	}
	events, err := h.Repo.GetEventsByGroupIDs(groupIDs)
	if err != nil {
		return internalServerError()
	}
	return c.JSON(http.StatusOK, service.FormatEventsRes(events))
}
