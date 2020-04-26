package router

import (
	"bytes"
	"net/http"
	repo "room/repository"
	"room/router/service"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/copier"
	"github.com/lestrrat-go/ical"

	"github.com/labstack/echo/v4"
)

// HandlePostEvent 部屋の使用宣言を作成
func (h *Handlers) HandlePostEvent(c echo.Context) error {
	var req service.EventReq
	if err := c.Bind(&req); err != nil {
		return badRequest(err, message(err.Error()))
	}
	eventParams := new(repo.WriteEventParams)
	err := copier.Copy(&eventParams, req)
	if err != nil {
		return internalServerError(err)
	}

	eventParams.CreatedBy, _ = getRequestUserID(c)

	event, err := h.Repo.CreateEvent(*eventParams)
	if err != nil {
		return judgeErrorResponse(err)
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
			return judgeErrorResponse(err)
		}
	}

	return c.JSON(http.StatusCreated, service.FormatEventRes(event))
}

// HandleGetEvent get one event
func (h *Handlers) HandleGetEvent(c echo.Context) error {
	eventID, err := getPathEventID(c)
	if err != nil {
		return notFound(err)
	}

	event, err := h.Repo.GetEvent(eventID)
	if err != nil {
		return judgeErrorResponse(err)
	}
	res := service.FormatEventRes(event)
	return c.JSON(http.StatusOK, res)
}

// HandleGetEvents 部屋の使用宣言情報を取得
func (h *Handlers) HandleGetEvents(c echo.Context) error {
	values := c.QueryParams()

	start, end, err := getTiemRange(values)
	if err != nil {
		return badRequest(err, message("invalid time"))
	}

	events, err := h.Repo.GetAllEvents(&start, &end)
	if err != nil {
		return judgeErrorResponse(err)
	}
	res := service.FormatEventsRes(events)
	return c.JSON(http.StatusOK, res)
}

// HandleGetEventsByGroupID get events by groupID
// If groupID does not exist, this return []. Does not returns error.
func (h *Handlers) HandleGetEventsByGroupID(c echo.Context) error {
	groupID, err := getPathGroupID(c)
	if err != nil {
		return notFound(err)
	}
	events, err := h.Repo.GetEventsByGroupIDs([]uuid.UUID{groupID})
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, service.FormatEventsRes(events))

}

// HandleDeleteEvent 部屋の使用宣言を削除
func (h *Handlers) HandleDeleteEvent(c echo.Context) error {
	eventID, err := getPathEventID(c)
	if err != nil {
		return notFound(err)
	}

	if err = h.Repo.DeleteEvent(eventID); err != nil {
		return internalServerError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// HandleUpdateEvent 任意の要素を変更
func (h *Handlers) HandleUpdateEvent(c echo.Context) error {
	var req service.EventReq
	if err := c.Bind(&req); err != nil {
		return badRequest(err, message(err.Error()))
	}
	eventParams := new(repo.WriteEventParams)
	err := copier.Copy(&eventParams, req)
	if err != nil {
		return internalServerError(err)
	}
	eventID, err := getPathEventID(c)
	if err != nil {
		return notFound(err)
	}
	eventParams.CreatedBy, _ = getRequestUserID(c)

	event, err := h.Repo.UpdateEvent(eventID, *eventParams)
	if err != nil {
		return judgeErrorResponse(err)
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
		return judgeErrorResponse(err)
	}

	res := service.FormatEventRes(event)
	return c.JSON(http.StatusOK, res)
}

func (h *Handlers) HandleAddEventTag(c echo.Context) error {
	var req service.TagRelationReq
	if err := c.Bind(&req); err != nil {
		return badRequest(err)
	}
	eventID, err := getPathEventID(c)
	if err != nil {
		return notFound(err, message(err.Error()))
	}
	tag, err := h.Repo.CreateOrGetTag(req.Name)
	if err != nil {
		return judgeErrorResponse(err)
	}
	err = h.Repo.AddTagToEvent(eventID, tag.ID, false)
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandleDeleteEventTag(c echo.Context) error {
	eventID, err := getPathEventID(c)
	if err != nil {
		return notFound(err, message(err.Error()))
	}
	tagName := c.Param("tagName")
	tag, err := h.Repo.GetTagByName(tagName)
	if err != nil {
		return judgeErrorResponse(err)
	}
	err = h.Repo.DeleteTagInEvent(eventID, tag.ID, false)
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandleGetMeEvents(c echo.Context) error {
	userID, _ := getRequestUserID(c)

	token, _ := getRequestUserToken(c)
	res, err := h.Dao.GetEventsByUserID(token, userID)
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, res)

}

func (h *Handlers) HandleGetEventsByUserID(c echo.Context) error {
	userID, err := getPathUserID(c)
	if err != nil {
		return notFound(err)
	}

	token, _ := getRequestUserToken(c)
	res, err := h.Dao.GetEventsByUserID(token, userID)
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, res)
}

// HandleGetEventsByRoomID get events by roomID
// If roomID does not exist, this return []. Does not returns error.
func (h *Handlers) HandleGetEventsByRoomID(c echo.Context) error {
	roomID, _ := getPathRoomID(c)
	events, err := h.Repo.GetEventsByRoomIDs([]uuid.UUID{roomID})
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, service.FormatEventsRes(events))
}

func (h *Handlers) HandleGetEventActivities(c echo.Context) error {
	events, err := h.Repo.GetEventActivities(7)
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.JSON(http.StatusOK, service.FormatEventsRes(events))
}

func (h *Handlers) HandleGetiCalByPrivateID(c echo.Context) error {
	str := c.Param("secret")
	userID, err := uuid.FromString(str[:36])
	if err != nil {
		return notFound(err)
	}
	secret := str[36:]
	user, err := h.Dao.Repo.GetUser(userID)
	if err != nil {
		return judgeErrorResponse(err)
	}
	if user.ICalSecret != secret {
		return notFound(err)
	}
	cal, err := h.Dao.GetiCalByUserID(userID)
	if err != nil {
		return judgeErrorResponse(err)
	}
	var buf bytes.Buffer
	ical.NewEncoder(&buf).Encode(cal)

	return c.Blob(http.StatusOK, "text/calendar", buf.Bytes())
}
